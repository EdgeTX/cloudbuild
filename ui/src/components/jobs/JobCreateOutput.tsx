import { JobCreationStatus } from "@hooks/useCreateJobs";
import { Empty, List, Row, Spin } from "antd";
import JobCreateStatus from "./JobCreateStatus";

interface Props {
  jobsStatus: JobCreationStatus[];
  isLoading: boolean;
}

function JobCreateOutput({ jobsStatus, isLoading }: Props) {
  if (jobsStatus.length === 0) {
    return (
      <Row style={{ height: "100%" }} align={"middle"} justify={"center"}>
        {!isLoading && <Empty />}
        {isLoading && <Spin size="large" />}
      </Row>
    );
  }

  return (
    <List
      dataSource={jobsStatus}
      renderItem={(jobStatus) => (
        <List.Item>
          <JobCreateStatus {...{ jobStatus }} />
        </List.Item>
      )}
    />
  );
}

export default JobCreateOutput;
