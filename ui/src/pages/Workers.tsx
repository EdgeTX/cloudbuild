import { Col, Result, Row, Spin } from "antd";
import { useWorkersData } from "@hooks/useWorkersData";
import WorkerCard from "@comps/workers/WorkerCard";

function Workers() {
  const { isLoading, error, data } = useWorkersData();

  if (isLoading) {
    return (
      <Row justify={"center"} align={"middle"} style={{ height: "100%" }}>
        <Spin size="large" />
      </Row>
    );
  }

  if (error) {
    return <Result status="error" title={error.message} />;
  }

  return (
    <>
      <Row justify={"start"} gutter={[8, 8]}>
        {data?.map((worker) => (
          <Col key={worker.id} span="24" md={{ span: 8 }}>
            <WorkerCard worker={worker} />
          </Col>
        ))}
      </Row>
    </>
  );
}

export default Workers;
