import { useContext } from "react";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";
import { Flag, Job } from "@hooks/useJobsData";

interface JobCreationParams {
  target: string;
  release: string;
  flags: Flag[];
}

type JobCreationSuccess = {
  job: Pick<Job, "id" | "status">;
  params: JobCreationParams;
  error: never;
};

type JobCreationFail = {
  params: Partial<JobCreationParams>;
  error: string;
};

type JobCreationStatus = JobCreationSuccess | JobCreationFail;

const isSuccessful = (
  status: JobCreationStatus,
): status is JobCreationSuccess => {
  return !status.error;
};

async function sendJobCreationReq(token: string, job: JobCreationParams) {
  const res = await fetch("api/jobs", {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(job),
  });
  const data = await res.json();
  if (res.ok) return { job: data, params: job };
  return { ...data, params: job };
}

function useCreatejobs() {
  const { token } = useContext(AuthContext) as AuthContextType;

  const createJob = async (job: JobCreationParams) => {
    return (await sendJobCreationReq(token, job));
  };

  const createMultipleJobs = async (jobs: JobCreationParams[]) => {
    const results: JobCreationStatus[] = [];
    for (const job of jobs) {
      results.push(await sendJobCreationReq(token, job));
    }
    return results;
  };

  return { createJob, createMultipleJobs };
}

export type { JobCreationParams, JobCreationStatus };
export { isSuccessful, useCreatejobs };
