import { Card, Space, Typography } from "antd";
import { Worker } from "@hooks/useWorkersData";
import DateAgo from "@comps/DateAgo";

const { Text } = Typography;

interface Props {
  worker: Worker;
}

function WorkerCard({ worker }: Props) {
  return (
    <Card title={worker.hostname}>
      <Space>
        <Text type="secondary">Created:</Text>
        <DateAgo date={worker.created_at} />
      </Space>
      <br />
      <Space>
        <Text type="secondary">Updated:</Text>
        <DateAgo date={worker.updated_at} />
      </Space>
    </Card>
  );
}

export default WorkerCard;
