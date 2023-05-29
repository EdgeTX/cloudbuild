import { Badge, Button, Space, Tooltip } from "antd";
import { PresetStatusColorType } from "antd/es/_util/colors";
import { Job, JobStatus } from "@hooks/useJobsData";
import JobRemove from "@comps/jobs/JobRemove";
import JobFlags from "@comps/jobs/JobFlags";
import JobDuration from "@comps/jobs/JobDuration";
import DateAgo from "@comps/DateAgo";
import { useTargets } from "@/hooks/useTargets";
import { useEffect, useState } from "react";
import { ColumnsType } from "antd/es/table";
import { ColumnFilterItem } from "antd/es/table/interface";
import { MessageInstance } from "antd/es/message/interface";
import { CodeOutlined, EyeOutlined, FileTextOutlined } from "@ant-design/icons";
import { JobSelectionAction } from "./JobsTable";

const STATUS_MAP: Record<JobStatus, string> = {
  "VOID": "Created",
  "WAITING_FOR_BUILD": "Waiting",
  "BUILD_IN_PROGRESS": "Building",
  "BUILD_SUCCESS": "Success",
  "BUILD_ERROR": "Failed",
};

const BADGE_STATUS_MAP: Record<JobStatus, PresetStatusColorType> = {
  "VOID": "default",
  "WAITING_FOR_BUILD": "warning",
  "BUILD_IN_PROGRESS": "processing",
  "BUILD_SUCCESS": "success",
  "BUILD_ERROR": "error",
};

function mapFilters(values: string[]) {
  return values.map((value) => ({ text: value, value: value }));
}

interface Params {
  setSelectedJob: React.Dispatch<
    React.SetStateAction<JobSelectionAction | undefined>
  >;
  messageApi: MessageInstance;
}

function useColumns({ setSelectedJob, messageApi }: Params) {
  const { targets } = useTargets();
  const [targetFilters, setTargetFilters] = useState<ColumnFilterItem[]>([]);
  const [releaseFilters, setReleaseFilters] = useState<ColumnFilterItem[]>([]);

  useEffect(() => {
    if (!targets) return;
    setReleaseFilters(mapFilters(Object.keys(targets.releases).sort()));
    setTargetFilters(mapFilters(Object.keys(targets.targets).sort()));
  }, [targets]);

  const columns: ColumnsType<Job> = [
    {
      title: "Created",
      key: "created_at",
      dataIndex: "created_at",
      render: (date: string) => <DateAgo date={date} />,
      sorter: true,
      defaultSortOrder: "descend",
    },
    {
      title: "Status",
      dataIndex: "status",
      render: (jobStatus: JobStatus, job: Job) => {
        let status = BADGE_STATUS_MAP[jobStatus];
        if (jobStatus === "BUILD_SUCCESS" && job.build_attempts > 1) {
          status = "warning";
        }
        return <Badge status={status} text={STATUS_MAP[jobStatus]} />;
      },
    },
    {
      title: "Duration",
      key: "duration",
      render: (job: Job) => <JobDuration job={job} />,
      sorter: true,
    },
    {
      title: "Target",
      dataIndex: "target",
      filters: targetFilters,
      filterSearch: true,
    },
    {
      title: "Release",
      dataIndex: "release",
      filters: releaseFilters,
      filterSearch: true,
    },
    {
      title: "Flags",
      render: (job: Job) => <JobFlags flags={job.flags} />,
    },
    {
      title: "Action",
      render: (job: Job) => (
        <Space size={0} wrap>
          <Tooltip title="detail">
            <Button
              type="link"
              icon={<EyeOutlined />}
              onClick={() => setSelectedJob({ job, action: "detail" })}
            />
          </Tooltip>
          <Tooltip title="json">
            <Button
              type="link"
              icon={<FileTextOutlined />}
              onClick={() => setSelectedJob({ job, action: "json" })}
            />
          </Tooltip>
          <Tooltip title="logs">
            <Button
              type="link"
              icon={<CodeOutlined />}
              onClick={() => setSelectedJob({ job, action: "logs" })}
            />
          </Tooltip>
          <JobRemove job={job} messageApi={messageApi} />
        </Space>
      ),
    },
  ];

  return { columns };
}

export { useColumns, STATUS_MAP };
