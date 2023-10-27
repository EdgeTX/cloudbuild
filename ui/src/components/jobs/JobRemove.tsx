import { Button, Popconfirm } from "antd";
import { useContext } from "react";
import { AuthContext, AuthContextType } from "@hooks/useAuthenticated";
import { Job } from "@hooks/useJobsData";
import { useQueryClient } from "@tanstack/react-query";
import { MessageInstance } from "antd/es/message/interface";
import { DeleteOutlined } from "@ant-design/icons";

function sendJobDeleteRequest(token: string, id: string) {
  return fetch("api/job/" + id, {
    method: "DELETE",
    headers: {
      "Authorization": `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });
}

interface Props {
  job: Job;
  messageApi: MessageInstance;
}

function JobRemove({ job, messageApi }: Props) {
  const { token } = useContext(AuthContext) as AuthContextType;
  const queryClient = useQueryClient();

  if (
    job.status === "BUILD_IN_PROGRESS" || job.status === "WAITING_FOR_BUILD"
  ) return null;

  const remove = async () => {
    messageApi.loading("Remove action");
    const res = await sendJobDeleteRequest(token, job.id);
    messageApi.destroy();

    if (!res.ok) {
      messageApi.error("Error while removing job");
      return;
    }

    queryClient.invalidateQueries({ queryKey: ["jobs"] });
    messageApi.success("Job successfuly removed");
  };

  return (
    <Popconfirm
      title="Delete the job"
      description="Are you sure to delete this job?"
      onConfirm={remove}
      okText="Yes"
      cancelText="No"
    >
      <Button type="link" icon={<DeleteOutlined />} />
    </Popconfirm>
  );
}

export default JobRemove;
