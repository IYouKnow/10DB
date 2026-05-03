import { MouseEvent as ReactMouseEvent, useEffect, useMemo, useRef, useState } from "react";
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

function createTableAt(position: { x: number; y: number }): TableBlueprint {
  const table = createTable();
  return {
    ...table,
    position
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
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; boardX: number; boardY: number } | null>(null);
  const boardRef = useRef<HTMLDivElement | null>(null);

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

  useEffect(() => {
    function closeMenu() {
      setContextMenu(null);
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setContextMenu(null);
      }
    }

    window.addEventListener("click", closeMenu);
    window.addEventListener("keydown", onKeyDown);
    return () => {
      window.removeEventListener("click", closeMenu);
      window.removeEventListener("keydown", onKeyDown);
    };
  }, []);

  function updateTable(nextTable: TableBlueprint) {
    onChange({
      ...blueprint,
      tables: blueprint.tables.map((table) => (table.id === nextTable.id ? nextTable : table))
    });
  }

  function addTable(position = { x: 120, y: 120 }) {
    const table = createTableAt(position);
    onChange({ ...blueprint, tables: [...blueprint.tables, table] });
    setSelectedTableId(table.id);
    return table;
  }

  function openContextMenu(event: ReactMouseEvent<HTMLDivElement>) {
    event.preventDefault();
    const rect = boardRef.current?.getBoundingClientRect();
    const localX = rect ? event.clientX - rect.left : 180;
    const localY = rect ? event.clientY - rect.top : 180;
    const boardX = Math.max(24, localX - 80);
    const boardY = Math.max(24, localY - 40);
    setContextMenu({
      x: localX,
      y: localY,
      boardX,
      boardY
    });
  }

  return (
    <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_360px]">
      <Card className="relative min-h-[calc(100vh-220px)] overflow-hidden border-none bg-[radial-gradient(circle_at_top_left,_rgba(141,210,200,0.18),_transparent_28%),linear-gradient(180deg,_#07111f_0%,_#11223a_100%)] p-0 text-white shadow-[0_30px_90px_rgba(8,17,31,0.28)]">
        <div className="pointer-events-none absolute inset-x-0 top-0 z-10 flex items-start justify-between px-6 py-6">
          <div className="max-w-xl">
            <p className="text-xs font-semibold uppercase tracking-[0.25em] text-glow/80">Schema board first</p>
            <h2 className="mt-2 text-3xl font-black tracking-tight">Design the database visually</h2>
            <p className="mt-2 text-sm text-slate-300">
              Right-click anywhere on the board to open actions like adding a PostgreSQL table. The connection details stay secondary.
            </p>
          </div>
          <div className="pointer-events-auto flex flex-wrap justify-end gap-2">
            <Button className="bg-white/12 text-white hover:bg-white/18" onClick={() => addTable()}>
              Add table
            </Button>
            <Button className="bg-glow text-ink hover:bg-glow/80" onClick={onSave}>
              Save blueprint
            </Button>
            <Button className="bg-ember text-ink hover:bg-ember/90" onClick={onPreview}>
              Preview SQL
            </Button>
          </div>
        </div>

        <div ref={boardRef} className="h-[calc(100vh-220px)] min-h-[760px]" onContextMenu={openContextMenu}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            fitView
            colorMode="dark"
            onPaneClick={() => setContextMenu(null)}
            onNodeClick={(_, node) => {
              setSelectedTableId(node.id);
              setContextMenu(null);
            }}
            onNodeDragStop={(_, node) => {
              const table = blueprint.tables.find((item) => item.id === node.id);
              if (!table) return;
              updateTable({ ...table, position: node.position });
            }}
          >
            <MiniMap
              pannable
              zoomable
              className="!bottom-6 !left-6 !rounded-2xl !border !border-white/10 !bg-slate-950/80"
              nodeColor={() => "#8dd2c8"}
            />
            <Controls className="!bottom-6 !right-6 !border !border-white/10 !bg-slate-950/70 !shadow-none" />
            <Background color="rgba(255,255,255,0.12)" gap={24} size={1.1} />
          </ReactFlow>
        </div>

        {contextMenu ? (
          <div
            className="absolute z-20 w-64 rounded-2xl border border-white/10 bg-slate-950/95 p-2 shadow-2xl backdrop-blur"
            style={{
              left: contextMenu.x,
              top: contextMenu.y
            }}
            onClick={(event) => event.stopPropagation()}
          >
            <button
              className="flex w-full items-center justify-between rounded-xl px-3 py-3 text-left text-sm text-white transition hover:bg-white/8"
              onClick={() => {
                addTable({ x: contextMenu.boardX, y: contextMenu.boardY });
                setContextMenu(null);
              }}
            >
              <span>Add PostgreSQL table</span>
              <span className="text-xs uppercase tracking-[0.18em] text-slate-400">create</span>
            </button>
            <button
              className="flex w-full items-center justify-between rounded-xl px-3 py-3 text-left text-sm text-white transition hover:bg-white/8"
              onClick={() => {
                onSave();
                setContextMenu(null);
              }}
            >
              <span>Save blueprint</span>
              <span className="text-xs uppercase tracking-[0.18em] text-slate-400">sync</span>
            </button>
            <button
              className="flex w-full items-center justify-between rounded-xl px-3 py-3 text-left text-sm text-white transition hover:bg-white/8"
              onClick={() => {
                onPreview();
                setContextMenu(null);
              }}
            >
              <span>Preview PostgreSQL SQL</span>
              <span className="text-xs uppercase tracking-[0.18em] text-slate-400">review</span>
            </button>
          </div>
        ) : null}
      </Card>

      <div className="space-y-5">
        <Card className="space-y-4 border-none bg-white/92 p-5 shadow-[0_18px_50px_rgba(8,17,31,0.14)]">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.22em] text-ember">Inspector</p>
              <h3 className="mt-1 text-xl font-black text-ink">Table editor</h3>
              <p className="text-sm text-slate-600">Select a board node to shape columns, defaults, and simple relations.</p>
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
            <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 p-5 text-sm text-slate-500">
              Right-click the board and choose <strong>Add PostgreSQL table</strong> to start modeling your schema.
            </div>
          )}
        </Card>

        <Card className="border-none bg-slate-950 text-white shadow-[0_18px_50px_rgba(8,17,31,0.18)]">
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-glow/80">Board workflow</p>
          <div className="mt-3 space-y-2 text-sm text-slate-300">
            <p>1. Right-click the board to add tables.</p>
            <p>2. Select a node to edit columns and relations.</p>
            <p>3. Save and preview SQL when the model looks right.</p>
          </div>
        </Card>
      </div>
    </div>
  );
}
