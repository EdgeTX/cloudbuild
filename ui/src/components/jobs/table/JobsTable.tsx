import { useState } from "react";
import { Modal, Result, Table, TablePaginationConfig, message } from "antd";
import { FilterValue, SorterResult } from "antd/es/table/interface";
import {
  Job,
  JobSortData,
  JobSortQuery,
  JobStatusQuery,
  useJobsData,
} from "@hooks/useJobsData";
import { useColumns } from "./useColumns";
import JobDetail from "@comps/jobs/JobDetail";

interface Props {
  style?: React.CSSProperties;
  status: JobStatusQuery;
}

function JobsTable({ style, status }: Props) {
  const [offset, setOffset] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  const [sortData, setSortData] = useState<JobSortData | undefined>({
    sort: "created_at",
    sort_desc: "true",
  });
  const [jobFilters, setJobFilters] = useState<
    Record<string, string[] | null>
  >();

  const { isLoading, error, data } = useJobsData(
    status,
    offset,
    pageSize,
    {
      sortData,
      target: jobFilters?.target?.join(","),
      release: jobFilters?.release?.join(","),
    },
  );

  const [selectedJob, setSelectedJob] = useState<Job | undefined>(undefined);
  const jobModalOpened = selectedJob != undefined;
  const hideJobModal = () => {
    setSelectedJob(undefined);
  };

  const [messageApi, contextHolder] = message.useMessage();
  const { columns } = useColumns({ setSelectedJob, messageApi });

  if (error) {
    return <Result status="error" title={error.message} />;
  }

  let paginationOption = {};
  if (data) {
    paginationOption = {
      current: Math.ceil((offset + 1) / data.limit),
      total: data.total_rows,
      pageSize: data.limit,
    };
  }

  const handleChange = (
    pagination: TablePaginationConfig,
    filters: Record<string, FilterValue | null>,
    sorter: SorterResult<any> | SorterResult<any>[],
  ) => {
    // offset and page size
    let newOffset = offset;
    if (pagination.current && pagination.pageSize) {
      newOffset = Math.ceil((pagination.current - 1) * pagination.pageSize);
      setPageSize(pagination.pageSize);
    }
    setOffset(newOffset);

    // sorting
    if (sorter instanceof Array) return;
    let newSortData = undefined;
    if (sorter.order && sorter.columnKey) {
      newSortData = {
        sort: sorter.columnKey as JobSortQuery,
        sort_desc: String(sorter.order === "descend"),
      } satisfies JobSortData;
    }
    setSortData(newSortData);

    // filters
    setJobFilters(filters as Record<string, string[] | null>);
  };

  return (
    <div style={style}>
      {contextHolder}
      <Table
        size={"middle"}
        loading={isLoading}
        dataSource={data?.rows}
        columns={columns}
        rowKey={"id"}
        onChange={handleChange}
        pagination={{ ...paginationOption, size: "default" }}
      />
      <Modal
        open={jobModalOpened}
        onCancel={hideJobModal}
        centered
        width={800}
        footer={[]}
      >
        {selectedJob && <JobDetail job={selectedJob} />}
      </Modal>
    </div>
  );
}

export default JobsTable;
