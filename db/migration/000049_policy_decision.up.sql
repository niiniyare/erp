CREATE TABLE policy_decisions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  policy_evaluation_id UUID NOT NULL REFERENCES policy_evaluations(id) ON DELETE CASCADE,
  policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
  decision VARCHAR(20) NOT NULL,
  reason TEXT,
  matched_rule TEXT,
  evaluation_ms BIGINT,
  target_matched BOOLEAN,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT fk_policy_evaluation FOREIGN KEY (policy_evaluation_id) REFERENCES policy_evaluations(id),
  CONSTRAINT fk_policy FOREIGN KEY (policy_id) REFERENCES policies(id)
);

CREATE INDEX idx_policy_decisions_evaluation_id ON policy_decisions(policy_evaluation_id);

CREATE INDEX idx_policy_decisions_policy_id ON policy_decisions(policy_id);

COMMENT ON TABLE policy_decisions IS 'Stores individual policy decisions made during a policy evaluation.';
