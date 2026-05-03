export const mockTables = [
  {
    id: 'tbl_users',
    name: 'users',
    x: 60,
    y: 80,
    columns: [
      { id: 'c1', name: 'id', type: 'uuid', primaryKey: true, nullable: false, unique: true, defaultValue: 'gen_random_uuid()' },
      { id: 'c2', name: 'name', type: 'text', primaryKey: false, nullable: false, unique: false, defaultValue: '' },
      { id: 'c3', name: 'email', type: 'text', primaryKey: false, nullable: false, unique: true, defaultValue: '' },
      { id: 'c4', name: 'created_at', type: 'timestamp', primaryKey: false, nullable: true, unique: false, defaultValue: 'NOW()' },
    ],
  },
  {
    id: 'tbl_posts',
    name: 'posts',
    x: 420,
    y: 60,
    columns: [
      { id: 'c5', name: 'id', type: 'uuid', primaryKey: true, nullable: false, unique: true, defaultValue: 'gen_random_uuid()' },
      { id: 'c6', name: 'user_id', type: 'uuid', primaryKey: false, nullable: false, unique: false, defaultValue: '', fk: { table: 'users', column: 'id' } },
      { id: 'c7', name: 'title', type: 'text', primaryKey: false, nullable: false, unique: false, defaultValue: '' },
      { id: 'c8', name: 'content', type: 'text', primaryKey: false, nullable: true, unique: false, defaultValue: '' },
      { id: 'c9', name: 'published', type: 'boolean', primaryKey: false, nullable: false, unique: false, defaultValue: 'false' },
      { id: 'c10', name: 'created_at', type: 'timestamp', primaryKey: false, nullable: true, unique: false, defaultValue: 'NOW()' },
    ],
  },
  {
    id: 'tbl_comments',
    name: 'comments',
    x: 420,
    y: 380,
    columns: [
      { id: 'c11', name: 'id', type: 'uuid', primaryKey: true, nullable: false, unique: true, defaultValue: 'gen_random_uuid()' },
      { id: 'c12', name: 'post_id', type: 'uuid', primaryKey: false, nullable: false, unique: false, defaultValue: '', fk: { table: 'posts', column: 'id' } },
      { id: 'c13', name: 'user_id', type: 'uuid', primaryKey: false, nullable: false, unique: false, defaultValue: '', fk: { table: 'users', column: 'id' } },
      { id: 'c14', name: 'body', type: 'text', primaryKey: false, nullable: false, unique: false, defaultValue: '' },
      { id: 'c15', name: 'created_at', type: 'timestamp', primaryKey: false, nullable: true, unique: false, defaultValue: 'NOW()' },
    ],
  },
];

export const mockTableData = {
  users: {
    columns: ['id', 'name', 'email', 'created_at'],
    rows: [
      { id: 'a1b2c3d4-...', name: 'Alice Johnson', email: 'alice@example.com', created_at: '2026-04-10 09:23:00' },
      { id: 'e5f6g7h8-...', name: 'Bob Martinez', email: 'bob@example.com', created_at: '2026-04-11 14:55:00' },
      { id: 'i9j0k1l2-...', name: 'Carol White', email: 'carol@example.com', created_at: '2026-04-13 08:40:00' },
      { id: 'm3n4o5p6-...', name: 'David Kim', email: 'david@example.com', created_at: '2026-04-15 17:30:00' },
    ],
  },
  posts: {
    columns: ['id', 'user_id', 'title', 'published', 'created_at'],
    rows: [
      { id: 'q7r8s9t0-...', user_id: 'a1b2c3d4-...', title: 'Getting Started with PostgreSQL', published: true, created_at: '2026-04-12 10:00:00' },
      { id: 'u1v2w3x4-...', user_id: 'e5f6g7h8-...', title: 'Advanced SQL Patterns', published: false, created_at: '2026-04-14 11:30:00' },
      { id: 'y5z6a7b8-...', user_id: 'a1b2c3d4-...', title: 'Why Self-Hosting Wins', published: true, created_at: '2026-04-16 09:00:00' },
    ],
  },
  comments: {
    columns: ['id', 'post_id', 'user_id', 'body', 'created_at'],
    rows: [
      { id: 'c9d0e1f2-...', post_id: 'q7r8s9t0-...', user_id: 'e5f6g7h8-...', body: 'Great article!', created_at: '2026-04-12 12:00:00' },
      { id: 'g3h4i5j6-...', post_id: 'q7r8s9t0-...', user_id: 'i9j0k1l2-...', body: 'Very helpful, thanks.', created_at: '2026-04-12 15:20:00' },
    ],
  },
};

export const PG_TYPES = ['uuid', 'text', 'varchar', 'integer', 'decimal', 'boolean', 'timestamp', 'date', 'jsonb'];

export const TEMPLATES = [
  { id: 'blank', label: 'Blank', description: 'Start from scratch' },
  { id: 'blog', label: 'Blog', description: 'Users, posts, comments, tags' },
  { id: 'booking', label: 'Booking App', description: 'Services, slots, reservations' },
  { id: 'ecommerce', label: 'Ecommerce', description: 'Products, orders, customers' },
  { id: 'crm', label: 'CRM', description: 'Contacts, deals, activities' },
  { id: 'tasks', label: 'Task Manager', description: 'Projects, tasks, assignees' },
];
