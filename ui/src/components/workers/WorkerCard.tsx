import { Card, Descriptions, Progress, Space, Tooltip, Typography } from "antd";
import { Worker, WorkerDiagnostics } from "@hooks/useWorkersData";
import DateAgo from "@comps/DateAgo";

const { Text } = Typography;

interface Props {
  worker: Worker;
}

function formatBytes(bytes: number): string {
  const units = ["B", "KiB", "MiB", "GiB", "TiB"];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  return `${value.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}

// Highlights a value above what is expected on a healthy worker: at most one
// build (container, volume, source dir) in progress, no dangling image.
function Count({ value, max }: { value: number; max: number }) {
  return value > max ? <Text type="warning">{value}</Text> : <>{value}</>;
}

function DiskUsage({ diag }: { diag: WorkerDiagnostics }) {
  const used = diag.disk_total_bytes - diag.disk_free_bytes;
  const percent = diag.disk_total_bytes > 0
    ? Math.round((used / diag.disk_total_bytes) * 100)
    : 0;
  // Space used outside of the locations measured inside the worker
  // container: the VM OS, the worker image, Docker on the VM host, etc.
  const other = Math.max(
    0,
    used - diag.podman_storage_bytes - diag.tmp_bytes,
  );

  return (
    <>
      <Progress
        percent={percent}
        status={percent >= 90 ? "exception" : "normal"}
        strokeColor={percent >= 75 && percent < 90 ? "#faad14" : undefined}
        format={(p) => `${p}%`}
      />
      <Text type="secondary">
        {formatBytes(used)} used of {formatBytes(diag.disk_total_bytes)}
        {" "}({formatBytes(diag.disk_free_bytes)} free)
      </Text>
      <Descriptions
        size="small"
        column={1}
        style={{ marginTop: 8 }}
        items={[
          {
            key: "podman",
            label: "Podman storage",
            children: formatBytes(diag.podman_storage_bytes),
          },
          {
            key: "tmp",
            label: "/tmp",
            children: (
              <Space>
                {formatBytes(diag.tmp_bytes)}
                <Text type="secondary">
                  (<Count value={diag.source_dirs} max={1} /> source dirs)
                </Text>
              </Space>
            ),
          },
          {
            key: "other",
            label: (
              <Tooltip title="Used space not accounted for by podman storage or /tmp: VM system, worker image, Docker on the VM host...">
                Other
              </Tooltip>
            ),
            children: formatBytes(other),
          },
          {
            key: "objects",
            label: "Podman",
            children: (
              <Text>
                {diag.podman_images} images
                (<Count value={diag.podman_dangling_images} max={0} /> dangling),
                {" "}<Count value={diag.podman_containers} max={1} /> containers,
                {" "}<Count value={diag.podman_volumes} max={1} /> volumes
              </Text>
            ),
          },
          {
            key: "jobs",
            label: "Jobs",
            children: (
              <Space>
                {diag.jobs_processed}
                <Text type="secondary">since start</Text>
                <DateAgo date={diag.started_at} />
              </Space>
            ),
          },
          {
            key: "measured",
            label: "Measured",
            children: <DateAgo date={diag.collected_at} />,
          },
        ]}
      />
      {diag.errors?.map((err) => (
        <div key={err}>
          <Text type="danger">{err}</Text>
        </div>
      ))}
    </>
  );
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
      <div style={{ marginTop: 12 }}>
        {worker.diagnostics
          ? <DiskUsage diag={worker.diagnostics} />
          : <Text type="secondary">No diagnostics reported yet</Text>}
      </div>
    </Card>
  );
}

export default WorkerCard;
