import { JobCreationStatus, useCreatejobs } from "@hooks/useCreateJobs";
import { Flag } from "@hooks/useJobsData";
import { TargetFlag, Targets, useTargets } from "@hooks/useTargets";
import {
  InboxOutlined,
  MinusCircleOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import {
  Button,
  Divider,
  Form,
  FormListFieldData,
  message,
  Row,
  Select,
  Space,
  Upload,
} from "antd";
import { FormInstance, useForm } from "antd/es/form/Form";
import { useEffect, useState } from "react";
import { DefaultOptionType } from "antd/es/select";

function mapToSelect(values: string[]) {
  return values.map((value) => ({ value, label: value }));
}

function getAdditionalTags(targets: Targets, target: string) {
  let flags: Record<string, TargetFlag> = {};

  const tags = targets.targets[target]?.tags;
  if (!tags) return flags;

  flags = tags.reduce((obj, tag) => {
    obj = { ...targets.tags[tag].flags };
    return obj;
  }, {});

  return flags;
}

interface FormTagProps {
  form: FormInstance<JobCreationParams>;
  value: FormListFieldData;
  index: number;
  remove: (index: number | number[]) => void;
  targets?: Targets;
  currentTarget: string;
}

function FormTag(
  { form, value, index, remove, targets, currentTarget }: FormTagProps,
) {
  const [currentFlag, setCurrentFlag] = useState("");
  const [flagValues, setFlagValues] = useState<DefaultOptionType[]>([]);
  const [flagOptions, setFlagOptions] = useState<DefaultOptionType[]>([]);

  useEffect(() => {
    if (!targets) return;

    const additionalFlags = getAdditionalTags(targets, currentTarget);
    let values = [
      ...(targets.flags[currentFlag]?.values ?? []),
      ...(additionalFlags[currentFlag]?.values ?? []),
    ];
    values = [...new Set(values)];
    if (!values) return;

    // current value not in flag value? reset it
    const flagValues = form.getFieldsValue()?.flags ?? [];
    const currentValue = flagValues[index]?.value;
    if (currentValue && !values.includes(currentValue)) {
      flagValues[index].value = "";
    }

    setFlagValues(mapToSelect(values));
  }, [currentFlag, currentTarget, flagOptions, targets, form, index]);

  useEffect(() => {
    if (!targets) return;
    const flagKeys = Object.keys(targets.flags);
    setFlagOptions(mapToSelect(flagKeys));
  }, [targets]);

  return (
    <div style={{ display: "flex", alignItems: "baseline", gap: 10 }}>
      <Space.Compact block style={{ display: "flex" }}>
        <Form.Item
          style={{ flex: 1 }}
          name={[value.name, "name"]}
          rules={[{
            required: true,
            message: "Missing flag",
          }]}
        >
          <Select
            showSearch
            placeholder="Flag"
            onChange={(flag) => setCurrentFlag(flag)}
            options={flagOptions}
          />
        </Form.Item>

        <Form.Item
          style={{ flex: 1 }}
          name={[value.name, "value"]}
          rules={[{ required: true, message: "Missing value" }]}
        >
          <Select showSearch placeholder="Value" options={flagValues} />
        </Form.Item>
      </Space.Compact>
      <MinusCircleOutlined
        onClick={() => remove(value.name)}
        style={{ fontSize: "1.15rem", transform: "translate(0, 4px)" }}
      />
    </div>
  );
}

interface JobCreationParams {
  release: string;
  target: string;
  flags: Flag[];
}

function JobCreateForm() {
  const [form] = useForm<JobCreationParams>();
  const [messageApi, contextHolder] = message.useMessage();
  const { createJob, createMultipleJobs } = useCreatejobs(messageApi);

  // target and release options for form select input

  const { targets } = useTargets();
  const [releaseOptions, setReleaseOptions] = useState<DefaultOptionType[]>([]);
  const [targetOptions, setTargetOptions] = useState<DefaultOptionType[]>([]);
  const [currentTarget, setCurrentTarget] = useState("");
  const [currentRelease, setCurrentRelease] = useState("");

  useEffect(() => {
    if (!targets) return;
    const releasesKeys = Object.keys(targets.releases);

    setReleaseOptions(mapToSelect(releasesKeys));
    form.setFieldValue("release", releasesKeys[0]);
  }, [targets, form]);

  useEffect(() => {
    if (!targets) return;
    let releaseTargets = Object.keys(targets.targets);
    const excludeTargets = targets.releases[currentRelease]?.exclude_targets;

    if (excludeTargets) {
      releaseTargets = releaseTargets.filter((target) =>
        !excludeTargets.includes(target)
      );
    }

    setTargetOptions(mapToSelect(releaseTargets));
    form.setFieldValue("target", releaseTargets[0]);
  }, [targets, currentRelease, form]);

  // file list and file content for upload file form input

  const [fileList, setFileList] = useState<File[]>([]);
  const [jobsFileContent, setJobsFileContent] = useState<object | undefined>();

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
    let _result: JobCreationStatus[] = [];
    if (jobsFileContent) {
      _result = await createMultipleJobs(jobsFileContent as JobCreationParams[]);
    } else {
      _result = [await createJob(values)];
    }
   
  };

  return (
    <>
      {contextHolder}
      <Form
        labelCol={{ span: 4 }}
        aria-labelledby=""
        form={form}
        name="Job Creation"
        onFinish={onFinish}
        style={{ marginTop: "2rem" }}
      >
        <Form.Item
          label="Release"
          name="release"
          rules={[{ required: jobsFileContent == undefined }]}
        >
          <Select
            options={releaseOptions}
            onChange={(newRelease) => setCurrentRelease(newRelease)}
          />
        </Form.Item>

        <Form.Item
          label="Target"
          name="target"
          rules={[{ required: jobsFileContent == undefined }]}
        >
          <Select
            showSearch
            options={targetOptions}
            onChange={(newTarget) => setCurrentTarget(newTarget)}
          />
        </Form.Item>

        <Form.List name="flags" initialValue={[]}>
          {(fields, { add, remove }) => (
            <Form.Item label="Flags">
              {fields.map((value, index) => (
                <div key={value.key}>
                  <FormTag
                    {...{
                      form,
                      value,
                      index,
                      remove,
                      targets,
                      currentRelease,
                      currentTarget,
                    }}
                  />
                </div>
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
            </Form.Item>
          )}
        </Form.List>

        <Divider />

        <Form.Item
          labelCol={{ span: 24 }}
          style={{ flexDirection: "column" }}
          label="Upload multiple jobs"
        >
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

        <Row justify={"center"}>
          <Form.Item>
            <Button type="primary" htmlType="submit">
              Submit
            </Button>
          </Form.Item>
        </Row>
      </Form>
    </>
  );
}

export default JobCreateForm;
