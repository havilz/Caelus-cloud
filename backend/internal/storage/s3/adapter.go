package s3

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
	ProviderType    domain.StorageProviderType
}

type Adapter struct {
	client       *s3.Client
	presign      *s3.PresignClient
	providerType domain.StorageProviderType
	region       string
}

func NewAdapter(cfg Config) (*Adapter, error) {
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.ProviderType == "" {
		cfg.ProviderType = domain.StorageProviderS3
	}

	awsCfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	presignClient := s3.NewPresignClient(client)

	return &Adapter{
		client:       client,
		presign:      presignClient,
		providerType: cfg.ProviderType,
		region:       cfg.Region,
	}, nil
}

func (a *Adapter) CreateBucket(ctx context.Context, bucketName, region string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	if region == "" {
		region = a.region
	}

	if region != "" && region != "us-east-1" && a.providerType == domain.StorageProviderS3 {
		input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}

	_, err := a.client.CreateBucket(ctx, input)
	if err != nil {
		return a.mapError(err, fmt.Sprintf("failed to create bucket %s", bucketName))
	}

	return nil
}

func (a *Adapter) ListBuckets(ctx context.Context) ([]domain.Bucket, error) {
	output, err := a.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, a.mapError(err, "failed to list buckets")
	}

	buckets := make([]domain.Bucket, 0, len(output.Buckets))
	for _, b := range output.Buckets {
		createdAt := time.Now().UTC()
		if b.CreationDate != nil {
			createdAt = *b.CreationDate
		}

		bucketName := ""
		if b.Name != nil {
			bucketName = *b.Name
		}

		buckets = append(buckets, domain.Bucket{
			ID:           uuid.New(),
			Name:         bucketName,
			ProviderType: a.providerType,
			Region:       a.region,
			CreatedAt:    createdAt,
			UpdatedAt:    createdAt,
		})
	}

	return buckets, nil
}

func (a *Adapter) DeleteBucket(ctx context.Context, bucketName string) error {

	paginator := s3.NewListObjectsV2Paginator(a.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			break
		}
		if len(page.Contents) > 0 {
			var ids []s3types.ObjectIdentifier
			for _, obj := range page.Contents {
				ids = append(ids, s3types.ObjectIdentifier{Key: obj.Key})
			}
			_, _ = a.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &s3types.Delete{Objects: ids, Quiet: aws.Bool(true)},
			})
		}
	}

	_, err := a.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchBucket" || apiErr.ErrorCode() == "NotFound" || strings.Contains(apiErr.ErrorMessage(), "404") {
				return nil
			}
		}
		return a.mapError(err, fmt.Sprintf("failed to delete bucket %s", bucketName))
	}

	return nil
}

func (a *Adapter) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	_, err := a.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" || strings.Contains(apiErr.ErrorMessage(), "404") {
				return false, nil
			}
		}
		return false, a.mapError(err, fmt.Sprintf("failed to head bucket %s", bucketName))
	}

	return true, nil
}

func (a *Adapter) ListObjects(ctx context.Context, bucketName, prefix, delimiter string, maxKeys int32) ([]domain.ObjectItem, []string, error) {
	if maxKeys <= 0 || maxKeys > 1000 {
		maxKeys = 1000
	}

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int32(maxKeys),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	if delimiter != "" {
		input.Delimiter = aws.String(delimiter)
	}

	output, err := a.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, nil, a.mapError(err, fmt.Sprintf("failed to list objects in bucket %s", bucketName))
	}

	objects := make([]domain.ObjectItem, 0, len(output.Contents))
	for _, obj := range output.Contents {
		key := ""
		if obj.Key != nil {
			key = *obj.Key
		}

		etag := ""
		if obj.ETag != nil {
			etag = strings.Trim(*obj.ETag, "\"")
		}

		lastMod := time.Now().UTC()
		if obj.LastModified != nil {
			lastMod = *obj.LastModified
		}

		storageClass := string(obj.StorageClass)
		size := int64(0)
		if obj.Size != nil {
			size = *obj.Size
		}

		contentType := mime.TypeByExtension(filepath.Ext(key))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		objects = append(objects, domain.ObjectItem{
			Key:          key,
			Size:         size,
			ETag:         etag,
			ContentType:  contentType,
			LastModified: lastMod,
			StorageClass: storageClass,
		})
	}

	folders := make([]string, 0, len(output.CommonPrefixes))
	for _, cp := range output.CommonPrefixes {
		if cp.Prefix != nil {
			folders = append(folders, *cp.Prefix)
		}
	}

	return objects, folders, nil
}

func (a *Adapter) UploadObject(ctx context.Context, input domain.UploadObjectInput) (*domain.ObjectItem, error) {
	contentType := input.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(input.Key))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(input.BucketName),
		Key:         aws.String(input.Key),
		Body:        input.Body,
		ContentType: aws.String(contentType),
		Metadata:    input.Metadata,
	}

	if input.Size > 0 {
		putInput.ContentLength = aws.Int64(input.Size)
	}

	if input.StorageClass != "" {
		putInput.StorageClass = s3types.StorageClass(input.StorageClass)
	}

	output, err := a.client.PutObject(ctx, putInput)
	if err != nil {
		return nil, a.mapError(err, fmt.Sprintf("failed to upload object %s to bucket %s", input.Key, input.BucketName))
	}

	etag := ""
	if output.ETag != nil {
		etag = strings.Trim(*output.ETag, "\"")
	}

	return &domain.ObjectItem{
		Key:          input.Key,
		Size:         input.Size,
		ETag:         etag,
		ContentType:  contentType,
		LastModified: time.Now().UTC(),
		StorageClass: input.StorageClass,
		Metadata:     input.Metadata,
	}, nil
}

func (a *Adapter) DownloadObject(ctx context.Context, bucketName, key string) (*domain.ObjectContent, error) {
	output, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, a.mapError(err, fmt.Sprintf("failed to download object %s from bucket %s", key, bucketName))
	}

	etag := ""
	if output.ETag != nil {
		etag = strings.Trim(*output.ETag, "\"")
	}

	contentType := "application/octet-stream"
	if output.ContentType != nil && *output.ContentType != "" {
		contentType = *output.ContentType
	}

	lastMod := time.Now().UTC()
	if output.LastModified != nil {
		lastMod = *output.LastModified
	}

	contentLength := int64(0)
	if output.ContentLength != nil {
		contentLength = *output.ContentLength
	}

	return &domain.ObjectContent{
		Body:          output.Body,
		ContentLength: contentLength,
		ContentType:   contentType,
		ETag:          etag,
		LastModified:  lastMod,
	}, nil
}

func (a *Adapter) DeleteObject(ctx context.Context, bucketName, key string) error {
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return a.mapError(err, fmt.Sprintf("failed to delete object %s from bucket %s", key, bucketName))
	}

	return nil
}

func (a *Adapter) DeleteObjects(ctx context.Context, bucketName string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	objectIdentifiers := make([]s3types.ObjectIdentifier, 0, len(keys))
	for _, k := range keys {
		objectIdentifiers = append(objectIdentifiers, s3types.ObjectIdentifier{
			Key: aws.String(k),
		})
	}

	_, err := a.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &s3types.Delete{
			Objects: objectIdentifiers,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return a.mapError(err, fmt.Sprintf("failed to delete batch objects from bucket %s", bucketName))
	}

	return nil
}

func (a *Adapter) GetObjectMetadata(ctx context.Context, bucketName, key string) (*domain.ObjectItem, error) {
	output, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, a.mapError(err, fmt.Sprintf("failed to get metadata for object %s in bucket %s", key, bucketName))
	}

	etag := ""
	if output.ETag != nil {
		etag = strings.Trim(*output.ETag, "\"")
	}

	contentType := "application/octet-stream"
	if output.ContentType != nil && *output.ContentType != "" {
		contentType = *output.ContentType
	}

	lastMod := time.Now().UTC()
	if output.LastModified != nil {
		lastMod = *output.LastModified
	}

	size := int64(0)
	if output.ContentLength != nil {
		size = *output.ContentLength
	}

	return &domain.ObjectItem{
		Key:          key,
		Size:         size,
		ETag:         etag,
		ContentType:  contentType,
		LastModified: lastMod,
		StorageClass: string(output.StorageClass),
		Metadata:     output.Metadata,
	}, nil
}

func (a *Adapter) GenerateSignedURL(ctx context.Context, bucketName, key string, operation domain.SignedURLOperation, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	switch operation {
	case domain.SignedURLOpDownload:
		req, err := a.presign.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = expiry
		})
		if err != nil {
			return "", a.mapError(err, fmt.Sprintf("failed to presign download URL for %s", key))
		}
		return req.URL, nil

	case domain.SignedURLOpUpload:
		req, err := a.presign.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = expiry
		})
		if err != nil {
			return "", a.mapError(err, fmt.Sprintf("failed to presign upload URL for %s", key))
		}
		return req.URL, nil

	default:
		return "", fmt.Errorf("%w: unsupported signed URL operation %s", domain.ErrValidation, operation)
	}
}

func (a *Adapter) mapError(err error, contextMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchBucket", "NoSuchKey", "NotFound":
			return fmt.Errorf("%w: %s (%s)", domain.ErrNotFound, contextMsg, apiErr.ErrorMessage())
		case "BucketAlreadyExists", "BucketAlreadyOwnedByYou":
			return fmt.Errorf("%w: %s (%s)", domain.ErrConflict, contextMsg, apiErr.ErrorMessage())
		case "AccessDenied", "InvalidAccessKeyId", "SignatureDoesNotMatch":
			return fmt.Errorf("%w: %s (%s)", domain.ErrForbidden, contextMsg, apiErr.ErrorMessage())
		default:
			return fmt.Errorf("%s: [%s] %s", contextMsg, apiErr.ErrorCode(), apiErr.ErrorMessage())
		}
	}

	return fmt.Errorf("%s: %w", contextMsg, err)
}
