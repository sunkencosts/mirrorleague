import { createContext, useContext, useMemo } from "react";
import { useAuthState } from "../hooks/useAuth";

type AuthContextValue = ReturnType<typeof useAuthState>;

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const { user, isLoading, userId } = useAuthState();
	const value = useMemo(() => ({ user, isLoading, userId }), [user, isLoading, userId]);
	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
	const ctx = useContext(AuthContext);
	if (!ctx) throw new Error("useAuth must be used inside AuthProvider");

	return ctx;
}
