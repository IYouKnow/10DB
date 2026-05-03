import React from 'react';

const TABLE_WIDTH = 224;
const TABLE_HEADER_H = 36;
const ROW_H = 32;

function getColumnY(table, colId) {
  const idx = table.columns.findIndex(c => c.id === colId);
  return table.y + TABLE_HEADER_H + idx * ROW_H + ROW_H / 2;
}

function getAnchor(fromTable, toTable) {
  const fromCX = fromTable.x + TABLE_WIDTH / 2;
  const toCX = toTable.x + TABLE_WIDTH / 2;
  return fromCX < toCX ? { from: 'right', to: 'left' } : { from: 'left', to: 'right' };
}

export default function RelationLines({ tables }) {
  const lines = [];

  tables.forEach((table) => {
    table.columns.forEach((col) => {
      if (!col.fk) return;
      const toTable = tables.find(t => t.name === col.fk.table);
      if (!toTable) return;

      const toCol = toTable.columns.find(c => c.name === col.fk.column);
      if (!toCol) return;

      const anchor = getAnchor(table, toTable);
      const x1 = anchor.from === 'right' ? table.x + TABLE_WIDTH : table.x;
      const y1 = getColumnY(table, col.id);
      const x2 = anchor.to === 'left' ? toTable.x : toTable.x + TABLE_WIDTH;
      const y2 = getColumnY(toTable, toCol.id);
      const cx = (x1 + x2) / 2;

      lines.push({ key: `${col.id}-${toTable.id}`, x1, y1, x2, y2, cx });
    });
  });

  return (
    <svg className="absolute inset-0 w-full h-full pointer-events-none" style={{ overflow: 'visible' }}>
      <defs>
        <linearGradient id="relGrad" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" stopColor="hsl(199 89% 48% / 0.5)" />
          <stop offset="100%" stopColor="hsl(262 83% 58% / 0.5)" />
        </linearGradient>
        <marker id="arrowEnd" markerWidth="6" markerHeight="6" refX="5" refY="3" orient="auto">
          <path d="M0,0 L0,6 L6,3 z" fill="hsl(199 89% 48% / 0.6)" />
        </marker>
      </defs>
      {lines.map(({ key, x1, y1, x2, y2, cx }) => (
        <path
          key={key}
          d={`M ${x1} ${y1} C ${cx} ${y1}, ${cx} ${y2}, ${x2} ${y2}`}
          stroke="url(#relGrad)"
          strokeWidth="1.5"
          fill="none"
          strokeDasharray="6 4"
          markerEnd="url(#arrowEnd)"
        />
      ))}
    </svg>
  );
}