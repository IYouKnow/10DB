import { useMemo, useState } from "react";
import { useColumns, useRows, useTables } from "../../api/data";
import { Card } from "../../components/ui/Card";

export function DataBrowser({ projectId }: { projectId: string }) {
  const tables = useTables(projectId);
  const [tableName, setTableName] = useState("");
  const activeTable = useMemo(() => tableName || tables.data?.tables[0]?.name || "", [tableName, tables.data?.tables]);
  const columns = useColumns(projectId, activeTable);
  const rows = useRows(projectId, activeTable);

  return (
    <Card className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-bold text-ink">Data browser</h2>
          <p className="text-sm text-slate-600">Quickly inspect public tables without leaving the control panel.</p>
        </div>
        <select className="rounded-xl border border-slate-200 px-3 py-2 text-sm" value={activeTable} onChange={(event) => setTableName(event.target.value)}>
          {(tables.data?.tables ?? []).map((table) => (
            <option key={table.name} value={table.name}>
              {table.name}
            </option>
          ))}
        </select>
      </div>

      {activeTable ? (
        <>
          <div className="flex flex-wrap gap-2">
            {(columns.data?.columns ?? []).map((column) => (
              <span key={column.name} className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-700">
                {column.name} · {column.dataType}
              </span>
            ))}
          </div>
          <div className="overflow-x-auto rounded-2xl border border-slate-200">
            <table className="min-w-full text-left text-sm">
              <thead className="bg-slate-50">
                <tr>
                  {(rows.data?.columns ?? []).map((column) => (
                    <th key={column} className="px-3 py-2 font-semibold text-slate-700">
                      {column}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(rows.data?.rows ?? []).map((row, index) => (
                  <tr key={index} className="border-t border-slate-100">
                    {(rows.data?.columns ?? []).map((column) => (
                      <td key={column} className="px-3 py-2 align-top text-slate-600">
                        {typeof row[column] === "object" ? JSON.stringify(row[column]) : String(row[column] ?? "")}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      ) : (
        <p className="text-sm text-slate-500">Apply a schema first to browse tables and rows.</p>
      )}
    </Card>
  );
}
