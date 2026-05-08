import type React from "react";

export const PROFILE_FALLBACK =
	"data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'><rect width='24' height='24' fill='%231e3020'/><circle cx='12' cy='8' r='4' fill='%233a5c3a'/><path d='M5 20c0-4.2 3.5-7 7-7s7 2.8 7 7' fill='%233a5c3a'/></svg>";

export function onImageError(e: React.SyntheticEvent<HTMLImageElement>) {
	e.currentTarget.onerror = null;
	e.currentTarget.src = PROFILE_FALLBACK;
}
