import { useContext } from "react";
import { MessageInstance } from "antd/es/message/interface";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";
import { Flag, Job } from "@hooks/useJobsData";

interface JobCreationParams {
  target: string;
  release: string;
  flags: Flag[];
}

interface JobCreationSuccess {
  job: Job;
  error: never;
}

interface JobCreationFail {
  error: string;
}

type JobCreationStatus = JobCreationSuccess | JobCreationFail;

const isSuccessful = (
  status: JobCreationStatus,
): status is JobCreationSuccess => {
  return !!status.error;
};

function sendJobCreationReq(token: string, job: JobCreationParams) {
  return fetch("api/jobs", {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(job),
  });
}

function useCreatejobs(messageApi: MessageInstance) {
  const { token } = useContext(AuthContext) as AuthContextType;

  const startLoading = () => {
    messageApi.open({
      type: "loading",
      content: "Action in progress..",
      duration: 0,
    });
  };

  const stopLoading = () => {
    messageApi.destroy();
  };

  const createJob = async (job: JobCreationParams) => {
    startLoading();
    const response = await sendJobCreationReq(token, job);

    stopLoading();
    return await response.json();
  };

  const createMultipleJobs = async (jobs: JobCreationParams[]) => {
    const responses = [];
    startLoading();

    for (const job of jobs) {
      responses.push(await sendJobCreationReq(token, job));
    }

    const results: JobCreationStatus[] = [];
    for (const res of responses) {
      results.push(await res.json());
    }

    stopLoading();
    return results;
  };

  return { createJob, createMultipleJobs };
}

export type { isSuccessful, JobCreationParams, JobCreationStatus };
export { useCreatejobs };
