import { useQuery } from "@tanstack/react-query";
import { useContext } from "react";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";

const REFRESH_INTERVAL = 10000;

// Mirrors processor.Diagnostics on the server side.
interface WorkerDiagnostics {
  collected_at: string;
  started_at: string;
  jobs_processed: number;
  disk_total_bytes: number;
  disk_free_bytes: number;
  tmp_bytes: number;
  source_dirs: number;
  podman_storage_bytes: number;
  podman_images: number;
  podman_dangling_images: number;
  podman_containers: number;
  podman_volumes: number;
  errors?: string[];
}

interface Worker {
  id: string;
  hostname: string;
  created_at: string;
  updated_at: string;
  diagnostics: WorkerDiagnostics | null;
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

export type { Worker, WorkerDiagnostics };
export { useWorkersData };
