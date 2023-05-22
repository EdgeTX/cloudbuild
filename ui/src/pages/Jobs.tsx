import { Button, Row, Segmented } from "antd";
import JobsTable from "@comps/jobs/table/JobsTable";
import { JobStatusQuery } from "@hooks/useJobsData";
import { useState } from "react";
import { PlusOutlined } from "@ant-design/icons";
import JobCreation from "./JobCreation";

const JOB_STATUS_OPTIONS: Record<string, JobStatusQuery> = {
  "All": "all",
  "Successful": "success",
  "In progress": "in-progress",
  "Failed": "error",
};

function Jobs() {
  const [jobCreateOpen, setJobCreateOpen] = useState(false);
  const [status, setStatus] = useState<JobStatusQuery>("all");

  if (jobCreateOpen) {
    return <JobCreation {...{ setJobCreateOpen }} />;
  }

  return (
    <>
      <Row justify={"space-between"}>
        <Segmented
          size="middle"
          options={Object.keys(JOB_STATUS_OPTIONS)}
          onChange={(value) => setStatus(JOB_STATUS_OPTIONS[value])}
        />
        <Button
          shape="round"
          icon={<PlusOutlined />}
          onClick={() => setJobCreateOpen(true)}
        >
          Create
        </Button>
      </Row>
      <JobsTable
        status={status}
        style={{ marginTop: 10 }}
      />
    </>
  );
}

export default Jobs;
