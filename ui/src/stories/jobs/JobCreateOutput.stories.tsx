import type { Meta, StoryObj } from "@storybook/react";
import JobCreateOutput, { Props } from "@comps/jobs/JobCreateOutput";
import { withProvider } from "../decorators/withProvider";

const meta: Meta<Props> = {
  title: "jobs/JobCreateOutput",
  component: JobCreateOutput,
  decorators: [withProvider],
};

export default meta;
type Story = StoryObj<Props>;

export const Primary: Story = {
  args: {},
};
