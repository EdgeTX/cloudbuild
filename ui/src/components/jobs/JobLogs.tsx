import { LogsData, useJobLogs } from "@hooks/useJobLogs";
import { Job } from "@hooks/useJobsData";
import { Collapse, Empty, Result } from "antd";
import { STATUS_MAP } from "./table/useColumns";
import { CaretRightOutlined } from "@ant-design/icons";
import dayjs from "dayjs";
import { AnsiUp } from "ansi_up";

import "@/ansi-themes.css";
import { CollapsibleType } from "antd/es/collapse/CollapsePanel";

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

  data?.sort((a, b) => (a.created_at < b.created_at ? -1 : 1));
  const items = data?.map((logs, i) => ({
    label: STATUS_MAP[logs.from],
    key: logs.id,
    showArrow: logs.std_out.length > 0,
    collapsible:
      logs.std_out.length > 0 ? undefined : ("icon" as CollapsibleType),
    extra: getStepDuration(data, i),
    children: <TerminalOutput logs={logs.std_out} />,
  }));

  return (
    <Collapse
      items={items}
      expandIcon={({ isActive }) => (
        <CaretRightOutlined rotate={isActive ? 90 : 0} />
      )}
      style={{ maxHeight: "80vh", overflowY: "auto", marginRight: 20 }}
    />
  );
}

export default JobLogs;
