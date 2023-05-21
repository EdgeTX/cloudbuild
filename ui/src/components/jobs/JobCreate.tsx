import { JobCreationParams, useCreatejobs } from "@hooks/useCreateJobs";
import {
  InboxOutlined,
  MinusCircleOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import {
  Button,
  Divider,
  Form,
  Input,
  message,
  Modal,
  Space,
  Upload,
} from "antd";
import { useForm } from "antd/es/form/Form";
import { useState } from "react";

function JobCreate() {
  const [open, setOpen] = useState(false);
  const [form] = useForm<JobCreationParams>();
  const [messageApi, contextHolder] = message.useMessage();
  const { createJob, createMultipleJobs } = useCreatejobs(messageApi);

  const [fileList, setFileList] = useState<File[]>([]);
  const [jobsFileContent, setJobsFileContent] = useState<object | undefined>();

  const hideCreateModal = () => {
    form.resetFields();
    setFileList([]);
    setJobsFileContent(undefined);
    setOpen(false);
  };

  const onFileBeforeUpload = async (file: File) => {
    if (file.type === "application/json") {
      try {
        const jsonContent = JSON.parse(await file.text());
        setJobsFileContent(jsonContent);
        setFileList([file]);
      } catch (error) {
        messageApi.error("Could not parse JSON file!");
      }
    } else {
      messageApi.error("You can only upload JSON file!");
    }
    return false;
  };

  const onFileRemove = () => {
    setJobsFileContent(undefined);
    setFileList([]);
  };

  const onFinish = async (values: JobCreationParams) => {
    messageApi.open({
      type: "loading",
      content: "Action in progress..",
      duration: 0,
    });

    if (jobsFileContent) {
      createMultipleJobs(jobsFileContent as JobCreationParams[]);
    } else {
      createJob(values);
    }
  };

  return (
    <>
      {contextHolder}
      <Button disabled shape="round" onClick={() => setOpen(true)}>
        <PlusOutlined />
        Create
      </Button>
      <Modal
        title={"Create a job"}
        open={open}
        onOk={() => {
          form.submit();
        }}
        onCancel={hideCreateModal}
      >
        <Form
          form={form}
          name="Job Creation"
          onFinish={onFinish}
          style={{ marginTop: "2rem" }}
        >
          <Form.Item
            label="Commit hash"
            name="commit_hash"
            rules={[{ required: jobsFileContent == undefined }]}
          >
            <Input />
          </Form.Item>

          <Form.List name="flags" initialValue={[]}>
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space
                    key={key}
                    style={{ display: "flex", marginBottom: 8 }}
                    align="baseline"
                  >
                    <Form.Item
                      {...restField}
                      name={[name, "key"]}
                      rules={[{
                        required: true,
                        message: "Missing flag",
                      }]}
                    >
                      <Input placeholder="Flag" />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, "value"]}
                      rules={[{ required: true, message: "Missing value" }]}
                    >
                      <Input placeholder="Value" />
                    </Form.Item>
                    <MinusCircleOutlined onClick={() => remove(name)} />
                  </Space>
                ))}
                <Form.Item>
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    Add Flag
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>
          <Divider />

          <Form.Item label="Upload multiple jobs">
            <Upload.Dragger
              name="file"
              maxCount={1}
              fileList={fileList as any}
              onRemove={onFileRemove}
              beforeUpload={onFileBeforeUpload}
            >
              <p className="ant-upload-drag-icon">
                <InboxOutlined />
              </p>
              <p className="ant-upload-text">
                Click or drag file to this area to upload
              </p>
            </Upload.Dragger>
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

export default JobCreate;
