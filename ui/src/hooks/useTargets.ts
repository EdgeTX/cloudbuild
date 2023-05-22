import { useContext } from "react";
import { AuthContext, AuthContextType } from "./useAuthenticated";
import { useQuery } from "@tanstack/react-query";

interface Release {
  sha: string;
  exclude_targets?: string[];
}

interface Tag {
  flags: Record<string, string[]>;
}

interface Target {
  description: string;
}

interface Targets {
  releases: Record<string, Release>;
  flags: Record<string, string[]>;
  tags: Record<string, Tag>;
  targets: Record<string, Target>;
}

function useTargets() {
  const { token } = useContext(AuthContext) as AuthContextType;

  const { isLoading, error, data } = useQuery<Targets, Error>({
    queryKey: ["targets", { token }],
    queryFn: async () => {
      let response: Response;
      try {
        response = await fetch(
          "api/targets",
          {
            headers: {
              "Authorization": `Bearer ${token}`,
              "Content-Type": "application/json",
            },
          },
        );
      } catch (_err) {
        throw new Error("Oops! Could not fetch targets");
      }

      if (!response.ok) {
        throw new Error(
          "Oops! An error occurred while retrieving targets data",
        );
      }

      return await response.json();
    },
  });

  return { isLoading, error, targets: data};
}

export type { Release, Tag, Target, Targets };
export { useTargets };
