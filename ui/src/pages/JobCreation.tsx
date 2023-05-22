import { LeftOutlined } from "@ant-design/icons";
import { Button, Col, Divider, Row } from "antd";
import JobCreateForm from "@comps/jobs/JobCreateForm";
import JobCreateOutput from "@comps/jobs/JobCreateOutput";
import { Dispatch, SetStateAction } from "react";

interface Props {
  setJobCreateOpen: Dispatch<SetStateAction<boolean>>;
}

function JobCreation({ setJobCreateOpen }: Props) {
  return (
    <div style={{ height: "100%" }}>
      <Button onClick={() => setJobCreateOpen(false)} icon={<LeftOutlined />}>
        Back
      </Button>
      <Row style={{ height: "100%" }}>
        <Col flex={"1"}>
          <JobCreateForm />
        </Col>
        <Col>
          <Divider type="vertical" style={{ height: "100%" }} />
        </Col>
        <Col flex={"1"}>
          <JobCreateOutput />
        </Col>
      </Row>
    </div>
  );
}

export type { Props };
export default JobCreation;
