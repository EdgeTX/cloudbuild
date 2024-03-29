import Layout from "@comps/Layout";
import { Row, Segmented } from "antd";
import JobsTable from "@comps/jobs/table/JobsTable";
import { JobStatusQuery } from "@hooks/useJobsData";
import { useState } from "react";

const JOB_STATUS_OPTIONS: Record<string, JobStatusQuery> = {
  "All": "all",
  "Successful": "success",
  "In progress": "in-progress",
  "Failed": "error",
};

function Jobs() {
  const [status, setStatus] = useState<JobStatusQuery>("all");

  return (
    <Layout>
      <Row justify={"space-between"}>
        <Segmented
          size="middle"
          options={Object.keys(JOB_STATUS_OPTIONS)}
          onChange={(value) => setStatus(JOB_STATUS_OPTIONS[value])}
        />
      </Row>
      <JobsTable
        status={status}
        style={{ marginTop: 10 }}
      />
    </Layout>
  );
}

export default Jobs;
