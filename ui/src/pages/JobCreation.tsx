import Layout from "@comps/Layout";
import { Col, Divider, message, Row } from "antd";
import JobCreateForm from "@comps/jobs/JobCreateForm";
import JobCreateOutput from "@comps/jobs/JobCreateOutput";
import {
  JobCreationParams,
  JobCreationStatus,
  useCreatejobs,
} from "@/hooks/useCreateJobs";
import { useState } from "react";

function JobCreation() {
  const [messageApi, contextHolder] = message.useMessage();
  const { createJob, createMultipleJobs } = useCreatejobs();
  const [jobsStatus, setJobsStatus] = useState<JobCreationStatus[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const onFinish = async (
    values: JobCreationParams,
    jobsFileContent: object | undefined,
  ) => {
    setJobsStatus([]);
    setIsLoading(true);

    let results;
    if (jobsFileContent) {
      results = await createMultipleJobs(
        jobsFileContent as JobCreationParams[],
      );
    } else {
      results = [await createJob(values)];
    }

    setIsLoading(false);
    setJobsStatus(results);
  };

  return (
    <Layout fullHeight>
      {contextHolder}
      <div style={{ height: "100%" }}>
        <Row style={{ height: "100%" }}>
          <Col flex={"1"} style={{ overflowY: "auto", height: "100%" }}>
            <JobCreateForm {...{ messageApi, onFinish }} />
          </Col>
          <Col>
            <Divider type="vertical" style={{ height: "100%" }} />
          </Col>
          <Col flex={"1"} style={{ overflowY: "auto", height: "100%" }}>
            <JobCreateOutput {...{ jobsStatus, isLoading }} />
          </Col>
        </Row>
      </div>
    </Layout>
  );
}

export default JobCreation;
