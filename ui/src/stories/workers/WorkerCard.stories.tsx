import WorkerCard from "@comps/workers/WorkerCard";
import type { Meta, StoryObj } from "@storybook/react";

const meta: Meta = {
  title: "workers/WorkerCard",
  component: WorkerCard,
  decorators: [
    (Story) => (
      <div style={{ maxWidth: 420 }}>
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj;

const GiB = 1024 * 1024 * 1024;
const now = new Date();
const minutesAgo = (m: number) =>
  new Date(now.getTime() - m * 60 * 1000).toISOString();

const worker = {
  id: "3bea302e-3578-4949-abc9-8fac448ff9e3",
  hostname: "worker-1",
  created_at: minutesAgo(60 * 24 * 12),
  updated_at: minutesAgo(0.1),
};

export const Healthy: Story = {
  args: {
    worker: {
      ...worker,
      diagnostics: {
        collected_at: minutesAgo(2),
        started_at: minutesAgo(60 * 24 * 12),
        jobs_processed: 412,
        disk_total_bytes: 40 * GiB,
        disk_free_bytes: 29.5 * GiB,
        tmp_bytes: 0.4 * GiB,
        source_dirs: 1,
        podman_storage_bytes: 6.2 * GiB,
        podman_images: 4,
        podman_dangling_images: 0,
        podman_containers: 1,
        podman_volumes: 1,
      },
    },
  },
};

export const Leaking: Story = {
  args: {
    worker: {
      ...worker,
      diagnostics: {
        collected_at: minutesAgo(4),
        started_at: minutesAgo(60 * 24 * 55),
        jobs_processed: 2391,
        disk_total_bytes: 40 * GiB,
        disk_free_bytes: 2.1 * GiB,
        tmp_bytes: 21.7 * GiB,
        source_dirs: 57,
        podman_storage_bytes: 9.8 * GiB,
        podman_images: 7,
        podman_dangling_images: 3,
        podman_containers: 2,
        podman_volumes: 9,
        errors: ["du: exit status 1"],
      },
    },
  },
};

export const NoDiagnostics: Story = {
  args: {
    worker: { ...worker, diagnostics: null },
  },
};
