-- =====================================================
-- ROLLBACK USER ACTIVITIES TABLE MIGRATION
-- =====================================================
-- Drops user activities table and related objects

-- Drop utility functions
DROP FUNCTION IF EXISTS create_monthly_user_activities_partition(DATE);
DROP FUNCTION IF EXISTS drop_old_user_activities_partitions(INTEGER);

-- Drop all user activities partitions
-- DROP TABLE IF EXISTS user_activities_prev2;
-- DROP TABLE IF EXISTS user_activities_prev1;
-- DROP TABLE IF EXISTS user_activities_current;
-- DROP TABLE IF EXISTS user_activities_next1;
-- DROP TABLE IF EXISTS user_activities_next2;
--
-- Drop the main partitioned table (this will cascade to any remaining partitions)
DROP TABLE IF EXISTS user_activities;
