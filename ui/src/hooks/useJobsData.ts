import { useContext } from "react";
import { useQuery } from "@tanstack/react-query";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";

const REFRESH_INTERVAL = 10000;

type JobStatus =
  | "VOID"
  | "WAITING_FOR_BUILD"
  | "BUILD_IN_PROGRESS"
  | "BUILD_SUCCESS"
  | "BUILD_ERROR";

interface Artifact {
  id: string;
  slug: string;
  created_at: string;
  updated_at: string;
  download_url: string;
}

interface Flag {
  key: string;
  value: string;
}

interface Job {
  id: string;
  status: JobStatus;
  build_attempts: number;
  commit_hash: string;
  release: string;
  target: string;
  flags: Flag[];
  build_flags: Flag[];
  artifacts: Artifact[];
  container_image: string;
  build_flags_hash: string;
  build_started_at: string;
  build_ended_at: string;
  created_at: string;
  updated_at: string;
}

type JobStatusQuery =
  | "all"
  | "success"
  | "error"
  | "queued"
  | "building"
  | "in-progress";

type JobSortQuery =
  | "created_at"
  | "duration"
  | "updated_at"
  | "build_started_at"
  | "build_ended_at";

interface QueryParameter {
  sort?: JobSortQuery;
  sort_desc?: string;
  offset?: string;
  limit?: string;
  status?: JobStatusQuery;
  target?: string;
  release?: string;
}

interface JobSortData {
  sort: JobSortQuery;
  sort_desc: string;
}

interface JobsResponse {
  limit: number;
  rows: Job[];
  sort_desc: boolean;
  total_rows: number;
}

function useJobsData(
  statusFilter: JobStatusQuery,
  offset: number,
  pageSize: number,
  { sortData, target, release }: {
    sortData?: JobSortData;
    target?: string;
    release?: string;
  },
) {
  const { token } = useContext(AuthContext) as AuthContextType;

  const params = {
    ...sortData,
    status: statusFilter,
    offset: offset.toString(),
    limit: String(pageSize),
    target: target ?? "",
    release: release ?? "",
  } satisfies QueryParameter;

  const { isLoading, error, data } = useQuery<JobsResponse, Error>({
    refetchInterval: REFRESH_INTERVAL,
    keepPreviousData: true,
    queryKey: ["jobs", { params, token }],
    queryFn: async () => {
      let response: Response;
      try {
        response = await fetch(
          "api/jobs?" + new URLSearchParams(params),
          {
            headers: {
              "Authorization": `Bearer ${token}`,
              "Content-Type": "application/json",
            },
          },
        );
      } catch (_err) {
        throw new Error("Oops! Could not fetch jobs");
      }

      if (!response.ok) {
        throw new Error("Oops! An error occurred while retrieving job data");
      }

      return response.json();
    },
  });

  return { isLoading, error, data };
}

export type {
  Artifact,
  Flag,
  Job,
  JobSortData,
  JobSortQuery,
  JobsResponse,
  JobStatus,
  JobStatusQuery,
};
export { useJobsData };
