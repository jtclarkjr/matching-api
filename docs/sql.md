# Option 1: Create tables only

psql -d your_database -f internal/database/schema.sql

# Option 2: Create tables and indexes separately

psql -d your_database -f internal/database/schema.sql
psql -d your_database -f internal/database/indexes.sql

# Option 3: Create everything at once

psql -d your_database -f internal/database/full_schema.sql

# Clean slate (careful - deletes all data!)

psql -d your_database -f internal/database/drop_tables.sql
