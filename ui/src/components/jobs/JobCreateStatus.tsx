import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
} from "@ant-design/icons";
import { isSuccessful, JobCreationStatus } from "@hooks/useCreateJobs";
import { List, Space, Typography } from "antd";
import JobFlags from "./JobFlags";

const { Text } = Typography;

function StatusContainer(
  { children, title, icon }: {
    children: React.ReactNode;
    title: string;
    icon: React.ReactNode;
  },
) {
  return (
    <List.Item.Meta
      title={
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          {icon}
          <Text strong>{title}</Text>
        </div>
      }
      description={children}
    >
    </List.Item.Meta>
  );
}

function DescItem(
  { children, label }: { children: React.ReactNode; label: string },
) {
  return (
    <Space>
      <Text children={label + ":"} type="secondary" />
      {children}
    </Space>
  );
}

interface Props {
  jobStatus: JobCreationStatus;
}

function JobCreateStatus({ jobStatus }: Props) {
  const successful = isSuccessful(jobStatus);
  const release = jobStatus.params?.release;
  const target = jobStatus.params?.target;
  const flags = jobStatus.params?.flags;

  const paramsRender = (
    <Space size={"middle"}>
      <DescItem label="Release">
        <Text>{release ?? "unspecified"}</Text>
      </DescItem>
      <DescItem label="Target">
        <Text>{target ?? "unspecified"}</Text>
      </DescItem>
      <DescItem label="Flags">
        {flags && <JobFlags flags={flags} />}
        {(!flags || flags.length === 0) && <Text>none</Text>}
      </DescItem>
    </Space>
  );

  if (successful) {
    const alreadyBuilt = jobStatus.job.status === "BUILD_SUCCESS";
    const icon = alreadyBuilt
      ? (
        <ExclamationCircleOutlined
          style={{ fontSize: "1rem", color: "orange" }}
        />
      )
      : <CheckCircleOutlined style={{ fontSize: "1rem", color: "green" }} />;
    const title = alreadyBuilt ? "Already built" : "Success";

    return (
      <StatusContainer
        title={title}
        icon={icon}
      >
        {paramsRender}
        <br />
        <DescItem label="Status">
          <Text>{jobStatus.job.status}</Text>
        </DescItem>
      </StatusContainer>
    );
  }

  return (
    <StatusContainer
      title={"Failed"}
      icon={<CloseCircleOutlined style={{ fontSize: "1rem", color: "red" }} />}
    >
      {paramsRender}
      <br />
      <DescItem label="Error">
        <Text type="danger">{jobStatus.error}</Text>
      </DescItem>
    </StatusContainer>
  );
}

export default JobCreateStatus;
