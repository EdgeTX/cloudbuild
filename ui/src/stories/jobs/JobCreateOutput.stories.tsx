import type { Meta, StoryObj } from "@storybook/react";
import JobCreateOutput from "@comps/jobs/JobCreateOutput";
import { withProvider } from "../decorators/withProvider";

const meta: Meta = {
  title: "jobs/JobCreateOutput",
  component: JobCreateOutput,
  decorators: [withProvider],
};

export default meta;
type Story = StoryObj;

export const Primary: Story = {
  args: {},
};
