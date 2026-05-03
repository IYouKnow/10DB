import { useMemo, useState } from "react";
import { Background, Controls, Edge, MiniMap, Node, ReactFlow } from "@xyflow/react";
import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { Input } from "../../components/ui/Input";
import { COLUMN_TYPES } from "../../lib/constants";
import { ColumnBlueprint, SchemaBlueprint, TableBlueprint } from "../../lib/types";

function makeId(prefix: string) {
  return `${prefix}_${Math.random().toString(36).slice(2, 10)}`;
}

function createTable(): TableBlueprint {
  return {
    id: makeId("tbl"),
    name: "new_table",
    position: { x: 100, y: 100 },
    columns: [
      {
        id: makeId("col"),
        name: "id",
        type: "id",
        primaryKey: true,
        unique: true,
        nullable: false,
        default: null,
        config: {}
      }
    ],
    foreignKeys: []
  };
}

function createColumn(): ColumnBlueprint {
  return {
    id: makeId("col"),
    name: "new_column",
    type: "text",
    primaryKey: false,
    unique: false,
    nullable: true,
    default: null,
    config: {}
  };
}

export function SchemaBuilder({
  blueprint,
  onChange,
  onSave,
  onPreview,
  validationErrors
}: {
  blueprint: SchemaBlueprint;
  onChange: (next: SchemaBlueprint) => void;
  onSave: () => void;
  onPreview: () => void;
  validationErrors?: Record<string, string>;
}) {
  const [selectedTableId, setSelectedTableId] = useState<string | null>(blueprint.tables[0]?.id ?? null);

  const nodes = useMemo<Node[]>(
    () =>
      blueprint.tables.map((table) => ({
        id: table.id,
        position: table.position,
        data: { label: `${table.name}\n${table.columns.map((column) => `${column.name}: ${column.type}`).join("\n")}` }
      })),
    [blueprint.tables]
  );

  const edges = useMemo<Edge[]>(
    () =>
      blueprint.tables.flatMap((table) =>
        table.foreignKeys.map((fk) => ({
          id: fk.id,
          source: table.id,
          target: blueprint.tables.find((candidate) => candidate.name === fk.refTable)?.id ?? table.id,
          label: `${fk.columnNames[0]} -> ${fk.refTable}.${fk.refColumnNames[0]}`
        }))
      ),
    [blueprint.tables]
  );

  const selectedTable = blueprint.tables.find((table) => table.id === selectedTableId) ?? null;

  function updateTable(nextTable: TableBlueprint) {
    onChange({
      ...blueprint,
      tables: blueprint.tables.map((table) => (table.id === nextTable.id ? nextTable : table))
    });
  }

  return (
    <div className="grid gap-6 xl:grid-cols-[1.4fr_0.9fr]">
      <Card className="min-h-[620px] overflow-hidden p-0">
        <div className="flex items-center justify-between border-b border-slate-200 px-5 py-4">
          <div>
            <h2 className="text-lg font-bold text-ink">Visual schema board</h2>
            <p className="text-sm text-slate-600">Design tables visually, then save and preview generated PostgreSQL SQL.</p>
          </div>
          <div className="flex gap-2">
            <Button
              className="bg-slate-200 text-ink hover:bg-slate-300"
              onClick={() => {
                const table = createTable();
                onChange({ ...blueprint, tables: [...blueprint.tables, table] });
                setSelectedTableId(table.id);
              }}
            >
              Add table
            </Button>
            <Button className="bg-glow text-ink hover:bg-glow/80" onClick={onSave}>
              Save blueprint
            </Button>
            <Button onClick={onPreview}>Preview SQL</Button>
          </div>
        </div>
        <div className="h-[560px]">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            fitView
            onNodeClick={(_, node) => setSelectedTableId(node.id)}
            onNodeDragStop={(_, node) => {
              const table = blueprint.tables.find((item) => item.id === node.id);
              if (!table) return;
              updateTable({ ...table, position: node.position });
            }}
          >
            <MiniMap />
            <Controls />
            <Background gap={18} size={1} />
          </ReactFlow>
        </div>
      </Card>

      <Card className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-bold text-ink">Table editor</h3>
            <p className="text-sm text-slate-600">Pick a node and edit columns, defaults, and simple relations.</p>
          </div>
          {selectedTable ? (
            <button
              className="text-sm font-semibold text-rose-600"
              onClick={() => {
                onChange({ ...blueprint, tables: blueprint.tables.filter((table) => table.id !== selectedTable.id) });
                setSelectedTableId(blueprint.tables.find((table) => table.id !== selectedTable.id)?.id ?? null);
              }}
            >
              Delete table
            </button>
          ) : null}
        </div>

        <div className="flex flex-wrap gap-2">
          {blueprint.tables.map((table) => (
            <button
              key={table.id}
              className={`rounded-full px-3 py-1 text-sm font-semibold ${selectedTableId === table.id ? "bg-ink text-white" : "bg-slate-100 text-slate-700"}`}
              onClick={() => setSelectedTableId(table.id)}
            >
              {table.name}
            </button>
          ))}
        </div>

        {selectedTable ? (
          <div className="space-y-4">
            <label className="block text-sm font-medium text-slate-700">
              Table name
              <Input value={selectedTable.name} onChange={(event) => updateTable({ ...selectedTable, name: event.target.value.toLowerCase() })} />
            </label>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="font-semibold text-slate-800">Columns</h4>
                <Button
                  className="bg-slate-200 text-ink hover:bg-slate-300"
                  onClick={() => updateTable({ ...selectedTable, columns: [...selectedTable.columns, createColumn()] })}
                >
                  Add column
                </Button>
              </div>
              {selectedTable.columns.map((column) => (
                <div key={column.id} className="rounded-2xl border border-slate-200 p-3">
                  <div className="grid gap-2 md:grid-cols-2">
                    <Input
                      value={column.name}
                      onChange={(event) =>
                        updateTable({
                          ...selectedTable,
                          columns: selectedTable.columns.map((item) => (item.id === column.id ? { ...item, name: event.target.value.toLowerCase() } : item))
                        })
                      }
                    />
                    <select
                      className="rounded-xl border border-slate-200 px-3 py-2 text-sm"
                      value={column.type}
                      onChange={(event) =>
                        updateTable({
                          ...selectedTable,
                          columns: selectedTable.columns.map((item) => (item.id === column.id ? { ...item, type: event.target.value } : item))
                        })
                      }
                    >
                      {COLUMN_TYPES.map((type) => (
                        <option key={type} value={type}>
                          {type}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="mt-3 flex flex-wrap gap-4 text-sm text-slate-600">
                    {[
                      ["primaryKey", "Primary key"],
                      ["unique", "Unique"],
                      ["nullable", "Nullable"]
                    ].map(([key, label]) => (
                      <label key={key} className="flex items-center gap-2">
                        <input
                          checked={Boolean(column[key as keyof ColumnBlueprint])}
                          type="checkbox"
                          onChange={(event) =>
                            updateTable({
                              ...selectedTable,
                              columns: selectedTable.columns.map((item) =>
                                item.id === column.id ? { ...item, [key]: event.target.checked } : item
                              )
                            })
                          }
                        />
                        {label}
                      </label>
                    ))}
                  </div>

                  {column.type === "varchar" ? (
                    <div className="mt-3">
                      <Input
                        type="number"
                        placeholder="Length"
                        value={column.config.varcharLength ?? 255}
                        onChange={(event) =>
                          updateTable({
                            ...selectedTable,
                            columns: selectedTable.columns.map((item) =>
                              item.id === column.id
                                ? { ...item, config: { ...item.config, varcharLength: Number(event.target.value) } }
                                : item
                            )
                          })
                        }
                      />
                    </div>
                  ) : null}

                  {column.type === "decimal" ? (
                    <div className="mt-3 grid gap-2 md:grid-cols-2">
                      <Input
                        type="number"
                        placeholder="Precision"
                        value={column.config.decimalPrecision ?? 10}
                        onChange={(event) =>
                          updateTable({
                            ...selectedTable,
                            columns: selectedTable.columns.map((item) =>
                              item.id === column.id
                                ? { ...item, config: { ...item.config, decimalPrecision: Number(event.target.value) } }
                                : item
                            )
                          })
                        }
                      />
                      <Input
                        type="number"
                        placeholder="Scale"
                        value={column.config.decimalScale ?? 2}
                        onChange={(event) =>
                          updateTable({
                            ...selectedTable,
                            columns: selectedTable.columns.map((item) =>
                              item.id === column.id
                                ? { ...item, config: { ...item.config, decimalScale: Number(event.target.value) } }
                                : item
                            )
                          })
                        }
                      />
                    </div>
                  ) : null}

                  <div className="mt-3 flex justify-end">
                    <button
                      className="text-sm font-semibold text-rose-600"
                      onClick={() =>
                        updateTable({
                          ...selectedTable,
                          columns: selectedTable.columns.filter((item) => item.id !== column.id)
                        })
                      }
                    >
                      Remove column
                    </button>
                  </div>
                </div>
              ))}
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="font-semibold text-slate-800">Foreign keys</h4>
                <Button
                  className="bg-slate-200 text-ink hover:bg-slate-300"
                  onClick={() =>
                    updateTable({
                      ...selectedTable,
                      foreignKeys: [
                        ...selectedTable.foreignKeys,
                        {
                          id: makeId("fk"),
                          columnNames: [selectedTable.columns[0]?.name ?? "id"],
                          refTable: blueprint.tables[0]?.name ?? selectedTable.name,
                          refColumnNames: ["id"],
                          onDelete: "CASCADE",
                          onUpdate: "NO ACTION"
                        }
                      ]
                    })
                  }
                >
                  Add FK
                </Button>
              </div>
              {selectedTable.foreignKeys.map((fk) => (
                <div key={fk.id} className="grid gap-2 rounded-2xl border border-slate-200 p-3">
                  <select
                    className="rounded-xl border border-slate-200 px-3 py-2 text-sm"
                    value={fk.columnNames[0]}
                    onChange={(event) =>
                      updateTable({
                        ...selectedTable,
                        foreignKeys: selectedTable.foreignKeys.map((item) =>
                          item.id === fk.id ? { ...item, columnNames: [event.target.value] } : item
                        )
                      })
                    }
                  >
                    {selectedTable.columns.map((column) => (
                      <option key={column.id} value={column.name}>
                        {column.name}
                      </option>
                    ))}
                  </select>
                  <select
                    className="rounded-xl border border-slate-200 px-3 py-2 text-sm"
                    value={fk.refTable}
                    onChange={(event) =>
                      updateTable({
                        ...selectedTable,
                        foreignKeys: selectedTable.foreignKeys.map((item) =>
                          item.id === fk.id ? { ...item, refTable: event.target.value } : item
                        )
                      })
                    }
                  >
                    {blueprint.tables.map((table) => (
                      <option key={table.id} value={table.name}>
                        {table.name}
                      </option>
                    ))}
                  </select>
                  <select
                    className="rounded-xl border border-slate-200 px-3 py-2 text-sm"
                    value={fk.refColumnNames[0]}
                    onChange={(event) =>
                      updateTable({
                        ...selectedTable,
                        foreignKeys: selectedTable.foreignKeys.map((item) =>
                          item.id === fk.id ? { ...item, refColumnNames: [event.target.value] } : item
                        )
                      })
                    }
                  >
                    {(blueprint.tables.find((table) => table.name === fk.refTable)?.columns ?? []).map((column) => (
                      <option key={column.id} value={column.name}>
                        {column.name}
                      </option>
                    ))}
                  </select>
                </div>
              ))}
            </div>

            {validationErrors && Object.keys(validationErrors).length > 0 ? (
              <div className="rounded-2xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-700">
                {Object.entries(validationErrors).map(([key, value]) => (
                  <div key={key}>
                    <strong>{key}:</strong> {value}
                  </div>
                ))}
              </div>
            ) : null}
          </div>
        ) : (
          <p className="text-sm text-slate-500">Add a table to start modeling your schema.</p>
        )}
      </Card>
    </div>
  );
}
