import { useContext } from "react";
import { useQuery } from "@tanstack/react-query";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";
import { JobStatus } from "@hooks/useJobsData";

interface LogsData {
  id: string;
  from: JobStatus;
  to: JobStatus;
  created_at: string;
  updated_at: string;
  std_out: string;
}

function fetchLogs(id: string, token: string) {
  return fetch(`api/logs/${id}`, {
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });
}

function useJobLogs(id: string) {
  const { token } = useContext(AuthContext) as AuthContextType;

  const { error, isLoading, data } = useQuery<LogsData[], Error>({
    queryKey: ["logs", { id, token }],
    queryFn: async () => {
      let response: Response;
      try {
        response = await fetchLogs(id, token);
      } catch (_err) {
        throw new Error("Oops! Could not fetch logs");
      }
      if (!response.ok) {
        throw new Error("Oops! An error occurred while retrieving logs");
      }
      return response.json();
    },
  });

  return { error, isLoading, data }
}

export type {LogsData}
export { useJobLogs };
