import { createContext, useEffect, useState } from "react";

interface AuthContextType {
  checkAuth: () => void;
  token: string;
}

const AuthContext = createContext<AuthContextType | null>(null);

function useAuthenticated() {
  const [isAuth, setIsAuth] = useState<undefined | boolean>(undefined);
  const [authError, setAuthError] = useState("");
  const [token, setToken] = useState("");

  const checkAuth = () => {
    const userToken = localStorage.getItem("token");
    if (!userToken) {
      setIsAuth(false);
      setAuthError("No access token provided");
      return;
    }
    setIsAuth(true);
    setToken(userToken);
  };

  useEffect(() => {
    checkAuth();
  }, []);

  return { isAuth, authError, checkAuth, token: token };
}

export type { AuthContextType };
export { AuthContext, useAuthenticated };
