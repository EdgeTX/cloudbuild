import { Descriptions, Space, Typography } from "antd";
import { Job } from "@hooks/useJobsData";
import DateAgo from "@comps/DateAgo";
import JobFlags from "@comps/jobs/JobFlags";
import JobDuration from "@comps/jobs/JobDuration";
import Link from "antd/es/typography/Link";

const { Title, Text } = Typography;

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
        <Descriptions.Item label="Id">
          <Text copyable children={job.id} />
        </Descriptions.Item>
        <Descriptions.Item label="Commit hash">
          <Text copyable children={job.commit_hash} />
        </Descriptions.Item>
        <Descriptions.Item label="Release" children={job.release} />
        <Descriptions.Item label="Target" children={job.target} />
        <Descriptions.Item label="Status" children={job.status} />
        <Descriptions.Item label="Attempts" children={job.build_attempts} />
        <Descriptions.Item label="Flags">
          <JobFlags flags={job.flags} />
        </Descriptions.Item>
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
            <Link copyable href={artifact.download_url}>
              {artifact.download_url}
            </Link>
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
