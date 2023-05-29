import { Space, Tag } from "antd";
import { Flag } from "@hooks/useJobsData";

interface Props {
  flags: Flag[];
}

function JobFlags({ flags }: Props) {
  return (
    <Space size={0}>
      {flags.map((flag) => (
        <div key={flag.name}>
          <Tag color="purple">{flag.value}</Tag>
        </div>
      ))}
    </Space>
  );
}

export default JobFlags;
