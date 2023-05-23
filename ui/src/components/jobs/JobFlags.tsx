import { Tag } from "antd";
import { Job } from "@hooks/useJobsData";

interface Props {
  job: Job;
}

function JobFlags({ job }: Props) {
  return (
    <>
      {job.flags.map((flag) => (
        <div key={flag.name}>
          <Tag color="purple">{flag.value}</Tag>
        </div>
      ))}
    </>
  );
}

export default JobFlags;
