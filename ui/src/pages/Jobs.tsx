import { Layout, Row, Segmented } from "antd";
import JobsTable from "@comps/jobs/table/JobsTable";
import { JobStatusQuery } from "@hooks/useJobsData";
import { useState } from "react";
import JobCreate from "@/components/jobs/JobCreate";

const JOB_STATUS_OPTIONS: Record<string, JobStatusQuery> = {
  "All": "all",
  "Successful": "success",
  "In progress": "in-progress",
  "Failed": "error",
};

function Jobs() {
  const [status, setStatus] = useState<JobStatusQuery>("all");

  return (
    <>
      <Row justify={"space-between"}>
        <Segmented
          size="middle"
          options={Object.keys(JOB_STATUS_OPTIONS)}
          onChange={(value) => setStatus(JOB_STATUS_OPTIONS[value])}
        />
        <JobCreate />
      </Row>
      <JobsTable
        status={status}
        style={{ marginTop: 10 }}
      />
    </>
  );
}

export default Jobs;
