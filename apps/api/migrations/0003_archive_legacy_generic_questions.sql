UPDATE questions
SET status = 'ARCHIVED'
WHERE status <> 'ARCHIVED'
  AND (
    question_index IS NULL
    OR TRIM(question_index) = ''
    OR subtest_code IS NULL
    OR TRIM(subtest_code) = ''
  );
