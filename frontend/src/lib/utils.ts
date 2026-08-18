import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Menggabungkan beberapa class name CSS dengan resolusi konflik class Tailwind CSS.
 * @param inputs Daftar string, ekspresi kondisi, atau array class CSS.
 * @returns String class name yang telah digabungkan dan dibersihkan.
 */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
