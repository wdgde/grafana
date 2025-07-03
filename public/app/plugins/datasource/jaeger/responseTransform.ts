import {
  DataFrame,
  FieldType,
  MutableDataFrame,
  TraceSpanRow,
} from '@grafana/data';

import { Span, TraceProcess, TraceResponse } from './types';

export function createTraceFrame(data: TraceResponse): DataFrame {
  const spans = data.spans.map((s) => toSpanRow(s, data.processes));

  const frame = new MutableDataFrame({
    fields: [
      { name: 'traceID', type: FieldType.string },
      { name: 'spanID', type: FieldType.string },
      { name: 'parentSpanID', type: FieldType.string },
      { name: 'operationName', type: FieldType.string },
      { name: 'serviceName', type: FieldType.string },
      { name: 'serviceTags', type: FieldType.other },
      { name: 'startTime', type: FieldType.number },
      { name: 'duration', type: FieldType.number },
      { name: 'logs', type: FieldType.other },
      { name: 'references', type: FieldType.other, values: [] },
      { name: 'tags', type: FieldType.other },
      { name: 'warnings', type: FieldType.other },
      { name: 'stackTraces', type: FieldType.other },
    ],
    meta: {
      preferredVisualisationType: 'trace',
      custom: {
        traceFormat: 'jaeger',
      },
    },
  });

  for (const span of spans) {
    frame.add(span);
  }

  return frame;
}

function toSpanRow(span: Span, processes: Record<string, TraceProcess>): TraceSpanRow {
  const parentSpanID = span.references?.find((r) => r.refType === 'CHILD_OF')?.spanID;

  return {
    spanID: span.spanID,
    traceID: span.traceID,
    parentSpanID: parentSpanID,
    operationName: span.operationName,
    // from micro to millis
    startTime: span.startTime / 1000,
    duration: span.duration / 1000,
    logs: span.logs.map((l) => ({
      ...l,
      timestamp: l.timestamp / 1000,
    })),
    tags: span.tags,
    warnings: span.warnings ?? undefined,
    stackTraces: span.stackTraces,
    references: span.references?.filter((r) => r.spanID !== parentSpanID) ?? [], // parentSpanID is pushed to references in the transformTraceDataFrame method
    serviceName: processes[span.processID].serviceName,
    serviceTags: processes[span.processID].tags,
  };
}
