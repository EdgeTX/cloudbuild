import Layout from "@comps/Layout";
import { Col, Result, Row, Spin } from "antd";
import { useWorkersData } from "@hooks/useWorkersData";
import WorkerCard from "@comps/workers/WorkerCard";

function Workers() {
  const { isLoading, error, data } = useWorkersData();

  if (isLoading) {
    return (
      <Layout>
        <Row justify={"center"} align={"middle"} style={{ height: "100%" }}>
          <Spin size="large" />
        </Row>
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout>
        <Result status="error" title={error.message} />
      </Layout>
    );
  }

  return (
    <Layout>
      <Row justify={"start"} gutter={[8, 8]}>
        {data?.map((worker) => (
          <Col key={worker.id} span="24" md={{ span: 8 }}>
            <WorkerCard worker={worker} />
          </Col>
        ))}
      </Row>
    </Layout>
  );
}

export default Workers;
