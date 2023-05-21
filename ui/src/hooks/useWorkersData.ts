import { useQuery } from "@tanstack/react-query";
import { useContext } from "react";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";

const REFRESH_INTERVAL = 10000;

interface Worker {
  id: string;
  hostname: string;
  created_at: string;
  updated_at: string;
}

function useWorkersData() {
  const { token } = useContext(AuthContext) as AuthContextType;

  const { isLoading, error, data } = useQuery<Worker[], Error>({
    refetchInterval: REFRESH_INTERVAL,
    queryKey: ["workers", { token }],
    queryFn: async () => {
      let response: Response;
      try {
        response = await fetch(
          "api/workers",
          {
            headers: {
              "Authorization": `Bearer ${token}`,
              "Content-Type": "application/json",
            },
          },
        );
      } catch (_err) {
        throw new Error("Oops! Could not fetch workers");
      }

      if (!response.ok) {
        throw new Error(
          "Oops! An error occurred while retrieving workers data",
        );
      }

      const data: Worker[] = await response.json();
      data.sort((a: Worker, b: Worker) => a.hostname > b.hostname ? 1 : -1);
      return data;
    },
  });

  return { isLoading, error, data };
}

export type { Worker };
export { useWorkersData };
