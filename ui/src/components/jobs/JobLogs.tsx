import { LogsData, useJobLogs } from "@hooks/useJobLogs";
import { Job } from "@hooks/useJobsData";
import { Collapse, Empty, Result } from "antd";
import { STATUS_MAP } from "./table/useColumns";
import { CaretRightOutlined } from "@ant-design/icons";
import dayjs from "dayjs";
import { default as AnsiUp } from "ansi_up";

import "@/ansi-themes.css"

const { Panel } = Collapse;

function getStepDuration(data: LogsData[], index: number) {
  const current = data[index];
  const last = data[index - 1];
  if (!last) return "";

  const lhs = dayjs(last.created_at);
  const rhs = dayjs(current.created_at);
  const duration = dayjs.duration(rhs.diff(lhs));

  return `${duration.asSeconds().toFixed(1)} sec`;
}

interface Props {
  job: Job;
}

function TerminalOutput({ logs }: { logs: string }) {
  console.log(logs);
  const ansi_up = new AnsiUp();
  ansi_up.use_classes = true;
  return (
    <pre
      style={{ fontSize: "13px", fontFamily: "monospace" }}
      dangerouslySetInnerHTML={{ __html: ansi_up.ansi_to_html(logs) }}
    />
  );
}

function JobLogs({ job }: Props) {
  const { error, data } = useJobLogs(job.id);

  if (error) {
    return <Result status="error" title={error.message} />;
  }
  if (data?.length === 0) {
    return <Empty />;
  }

  return (
    <Collapse
      expandIcon={({ isActive }) => (
        <CaretRightOutlined rotate={isActive ? 90 : 0} />
      )}
      style={{ maxHeight: "80vh", overflowY: "auto", marginRight: 20 }}
    >
      {data &&
        data.map((logs, i) => (
          <Panel
            showArrow={logs.std_out.length > 0}
            collapsible={logs.std_out.length > 0 ? undefined : "icon"}
            header={STATUS_MAP[logs.from]}
            key={logs.id}
            extra={getStepDuration(data, i)}
          >
            <TerminalOutput logs={logs.std_out} />
          </Panel>
        ))}
    </Collapse>
  );
}

export default JobLogs;
