import { List, Tag } from "antd";
import { Job } from "@hooks/useJobsData";

interface Props {
  job: Job;
}

function JobFlags({ job }: Props) {
  return (
    <List
      dataSource={job.flags}
      renderItem={(flag) => (
        <Tag color="purple">{flag.value}</Tag>
      )}
    />
  );
}

export default JobFlags;
