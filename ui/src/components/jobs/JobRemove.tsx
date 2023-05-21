import { Popconfirm } from "antd";
import { useContext } from "react";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";
import { Job } from "@hooks/useJobsData";

interface Props {
  job: Job;
}

async function sendJobDeleteRequest(token: string, id: string) {
  const res = await fetch("api/job/" + id, {
    method: "DELETE",
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });
  console.log(await res.json());
}

function JobRemove({ job }: Props) {
  const { token } = useContext(AuthContext) as AuthContextType;

  if (
    job.status === "BUILD_IN_PROGRESS" || job.status === "WAITING_FOR_BUILD"
  ) return null;

  const remove = () => {
    sendJobDeleteRequest(token, job.id);
  };

  return (
    <Popconfirm
      title="Delete the job"
      description="Are you sure to delete this job?"
      onConfirm={remove}
      okText="Yes"
      cancelText="No"
    >
      <a>Delete</a>
    </Popconfirm>
  );
}

export default JobRemove;
