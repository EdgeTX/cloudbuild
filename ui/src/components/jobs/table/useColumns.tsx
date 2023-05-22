import { Badge, Space } from "antd";
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

const STATUS_MAP: Record<JobStatus, string> = {
  "VOID": "",
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

interface Params {
  setSelectedJob: React.Dispatch<React.SetStateAction<Job | undefined>>;
}

function mapFilters(values: string[]) {
  return values.map((value) => ({ text: value, value: value }));
}

function useColumns({ setSelectedJob }: Params) {
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
      render: (job: Job) => <JobFlags job={job} />,
    },
    {
      title: "Action",
      render: (job: Job) => (
        <Space size="middle" wrap>
          <a onClick={() => setSelectedJob(job)}>Detail</a>
          <JobRemove job={job} />
        </Space>
      ),
    },
  ];

  return { columns };
}

export { useColumns };
