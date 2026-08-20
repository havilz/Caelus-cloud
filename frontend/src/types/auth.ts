export interface User {
  id: string;
  email: string;
  full_name: string;
  is_active: boolean;
  created_at: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  role: "owner" | "admin" | "member" | "viewer";
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface AuthData {
  user: User;
  organization?: Organization;
  tokens: TokenPair;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput {
  full_name: string;
  email: string;
  password: string;
  organization_name?: string;
}
