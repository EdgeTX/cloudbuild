import { useContext } from "react";
import { MessageInstance } from "antd/es/message/interface";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";
import { Flag } from "@hooks/useJobsData";

interface JobCreationParams {
  commit_hash: string;
  flags: Flag[];
}

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

  const createJob = async (job: JobCreationParams) => {
    const response = await sendJobCreationReq(token, job);
    messageApi.destroy();

    if (response.ok) {
      messageApi.success(
        `Successfuly managed to send the job`
      );
      return undefined;
    }

    const error = await response.json();
    return error;
  };

  const createMultipleJobs = async (jobs: JobCreationParams[]) => {
    const responses = [];
    for (const job of jobs) {
      responses.push(await sendJobCreationReq(token, job));
    }

    const errors = [];
    let successfulNb = 0;

    for (const res of responses) {
      console.log(res.statusText);
      if (res.ok) {
        successfulNb += 1;
        continue;
      }
      errors.push(await res.json());
    }

    messageApi.destroy();

    if (errors.length === 0) {
      messageApi.success(
        `Successfuly managed to send the ${successfulNb} jobs`,
      );
      return undefined;
    }

    if (successfulNb !== 0) {
      messageApi.warning(
        `Managed to send ${successfulNb} out of ${responses.length} jobs`,
      );
    }

    return errors;
  };

  return { createJob, createMultipleJobs };
}

export type { JobCreationParams };
export { useCreatejobs };
