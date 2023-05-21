import { Descriptions, Space, Typography } from "antd";
import DescriptionsItem from "antd/es/descriptions/Item";
import { Job } from "@hooks/useJobsData";
import DateAgo from "@comps/DateAgo";
import JobFlags from "@comps/jobs/JobFlags";
import JobDuration from "@comps/jobs/JobDuration";

const { Title } = Typography;

interface DescriptionGroupProps {
  children: React.ReactNode;
  title: string;
  style?: React.CSSProperties;
}

function DescriptionGroup({ children, title, style }: DescriptionGroupProps) {
  return (
    <Space direction="vertical" size={5} style={style}>
      <Title level={5}>{title}</Title>
      <Descriptions size={"small"} bordered column={1}>
        {children}
      </Descriptions>
    </Space>
  );
}

interface Props {
  job: Job;
}

function JobDetail({ job }: Props) {
  const artifact = job.artifacts?.at(0);

  return (
    <Space direction="vertical">
      <DescriptionGroup title="Job">
        <DescriptionsItem label="Flags">
          <JobFlags job={job} />
        </DescriptionsItem>
        <Descriptions.Item label="Status">{job.status}</Descriptions.Item>
        <Descriptions.Item label="Target" children={job.target} />
        <Descriptions.Item label="Release" children={job.release} />
        <Descriptions.Item label="Commit hash" children={job.commit_hash} />
        <Descriptions.Item label="Attempts" children={job.build_attempts} />
        <Descriptions.Item label="Created">
          <DateAgo date={job.created_at} />
        </Descriptions.Item>
        <Descriptions.Item label="Updated">
          <DateAgo date={job.updated_at} />
        </Descriptions.Item>
        <Descriptions.Item label="duration">
          <JobDuration job={job} />
        </Descriptions.Item>
      </DescriptionGroup>
      {artifact && (
        <DescriptionGroup title="Artifact" style={{ marginTop: 10 }}>
          <Descriptions.Item label="Slug" children={artifact.slug} />
          <Descriptions.Item label="Download URL">
            <a href={artifact.download_url}>{artifact.download_url}</a>
          </Descriptions.Item>
          <Descriptions.Item label="Created">
            <DateAgo date={artifact.created_at} />
          </Descriptions.Item>
          <Descriptions.Item label="Updated">
            <DateAgo date={artifact.updated_at} />
          </Descriptions.Item>
        </DescriptionGroup>
      )}
    </Space>
  );
}

export default JobDetail;
