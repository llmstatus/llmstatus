#!/usr/bin/env node
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { LLMStatusClient } from "./client.js";
import {
  TOOL_NAME as LIST_PROVIDERS,
  TOOL_DESCRIPTION as LP_DESC,
  handleListProviders,
} from "./tools/list_providers.js";
import {
  TOOL_NAME as GET_PROVIDER_STATUS,
  TOOL_DESCRIPTION as GPS_DESC,
  TOOL_SCHEMA as GPS_SCHEMA,
  handleGetProviderStatus,
} from "./tools/get_provider_status.js";
import {
  TOOL_NAME as LIST_INCIDENTS,
  TOOL_DESCRIPTION as LI_DESC,
  TOOL_SCHEMA as LI_SCHEMA,
  handleListActiveIncidents,
} from "./tools/list_active_incidents.js";
import {
  TOOL_NAME as GET_INCIDENT,
  TOOL_DESCRIPTION as GI_DESC,
  TOOL_SCHEMA as GI_SCHEMA,
  handleGetIncidentDetail,
} from "./tools/get_incident_detail.js";
import {
  TOOL_NAME as GET_HISTORY,
  TOOL_DESCRIPTION as GH_DESC,
  TOOL_SCHEMA as GH_SCHEMA,
  handleGetProviderHistory,
} from "./tools/get_provider_history.js";
import {
  TOOL_NAME as COMPARE,
  TOOL_DESCRIPTION as CMP_DESC,
  TOOL_SCHEMA as CMP_SCHEMA,
  handleCompareProviders,
} from "./tools/compare_providers.js";
import { ApiError } from "./types.js";

const client = new LLMStatusClient();
const server = new McpServer({ name: "llmstatus", version: "1.0.0" });

function wrapError(err: unknown): string {
  if (err instanceof ApiError) return err.message;
  if (err instanceof Error) return err.message;
  return "An unexpected error occurred.";
}

server.tool(LIST_PROVIDERS, LP_DESC, {}, async () => ({
  content: [{ type: "text" as const, text: await handleListProviders(client).catch(wrapError) }],
}));

server.tool(GET_PROVIDER_STATUS, GPS_DESC, GPS_SCHEMA, async ({ id }) => ({
  content: [
    { type: "text" as const, text: await handleGetProviderStatus(id, client).catch(wrapError) },
  ],
}));

server.tool(LIST_INCIDENTS, LI_DESC, LI_SCHEMA, async ({ provider_id }) => ({
  content: [
    {
      type: "text" as const,
      text: await handleListActiveIncidents(provider_id, client).catch(wrapError),
    },
  ],
}));

server.tool(GET_INCIDENT, GI_DESC, GI_SCHEMA, async ({ id }) => ({
  content: [
    { type: "text" as const, text: await handleGetIncidentDetail(id, client).catch(wrapError) },
  ],
}));

server.tool(GET_HISTORY, GH_DESC, GH_SCHEMA, async ({ id, window }) => ({
  content: [
    {
      type: "text" as const,
      text: await handleGetProviderHistory(id, window ?? "30d", client).catch(wrapError),
    },
  ],
}));

server.tool(COMPARE, CMP_DESC, CMP_SCHEMA, async ({ ids }) => ({
  content: [
    { type: "text" as const, text: await handleCompareProviders(ids, client).catch(wrapError) },
  ],
}));

const transport = new StdioServerTransport();
await server.connect(transport);
