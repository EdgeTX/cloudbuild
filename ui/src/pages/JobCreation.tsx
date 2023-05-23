import { Col, Divider, Row } from "antd";
import JobCreateForm from "@comps/jobs/JobCreateForm";
import JobCreateOutput from "@comps/jobs/JobCreateOutput";

function JobCreation() {
  return (
    <div style={{ height: "100%" }}>
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
