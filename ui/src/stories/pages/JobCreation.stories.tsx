import type { Meta, StoryObj } from "@storybook/react";
import JobCreation from "@pages/JobCreation";
import { withProvider } from "../decorators/withProvider";

const meta: Meta = {
  title: "pages/JobCreation",
  component: JobCreation,
  decorators: [withProvider],
};

export default meta;
type Story = StoryObj;

export const Primary: Story = {
  args: {},
};
