import type { Meta, StoryObj } from "@storybook/react";
import JobCreateOutput from "@comps/jobs/JobCreateOutput";

const meta: Meta = {
  title: "jobs/JobCreateOutput",
  component: JobCreateOutput,
};

export default meta;
type Story = StoryObj;

export const Empty: Story = {
  args: {
    jobsStatus: [],
  },
};

export const Loading: Story = {
  args: {
    jobsStatus: [],
    isLoading: true,
  }
}

export const Errors: Story = {
  args: {
    jobsStatus: [
      {
        "error": "release not supported: ",
      },
      {
        "error": "release not supported: ",
      },
      {
        "error": "release not supported: ",
      },
      {
        "error": "release not supported: ",
      },
      {
        "error": "release not supported: ",
      },
      {
        "error": "release not supported: ",
      },
      {
        "error": "release not supported: ",
      },
      {
        "error": "release not supported: ",
      },
    ],
  },
};
