import { Job } from "@hooks/useJobsData";
import DateDuration from "@comps/DateDuration";

interface Props {
  job: Job;
}

function JobDuration({ job }: Props) {
  if (job.status === "BUILD_IN_PROGRESS") {
    return null;
  }

  return (
    <DateDuration
      lhsDate={job.created_at}
      rhsDate={job.updated_at}
    />
  );
}

export default JobDuration;
