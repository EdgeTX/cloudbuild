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
      lhsDate={job.build_started_at}
      rhsDate={job.build_ended_at}
    />
  );
}

export default JobDuration;
