#!/bin/bash
# extract-sections.sh

SOURCE="Readme_md_original.md"
DEST_BASE="schema"

# Extract Part I
sed -n '42,556p' $SOURCE >$DEST_BASE/01-foundations/01-introduction.md
sed -n '558,1089p' $SOURCE >$DEST_BASE/01-foundations/02-architecture.md
sed -n '1091,1633p' $SOURCE >$DEST_BASE/01-foundations/03-quick-start.md
sed -n '1635,2138p' $SOURCE >$DEST_BASE/01-foundations/04-concepts.md
sed -n '2140,2576p' $SOURCE >$DEST_BASE/01-foundations/05-comparison-amis.md

# Extract Part II
sed -n '2582,3760p' $SOURCE >$DEST_BASE/02-core-specification/06-schema-structure.md
sed -n '3762,5018p' $SOURCE >$DEST_BASE/02-core-specification/07-field-system.md
sed -n '5020,5813p' $SOURCE >$DEST_BASE/02-core-specification/08-validation.md
sed -n '5824,6752p' $SOURCE >$DEST_BASE/02-core-specification/09-layout.md
sed -n '6754,7336p' $SOURCE >$DEST_BASE/02-core-specification/10-conditional-visibility.md
sed -n '7338,7913p' $SOURCE >$DEST_BASE/02-core-specification/11-security-permissions.md
sed -n '7915,8452p' $SOURCE >$DEST_BASE/02-core-specification/12-events-actions.md
sed -n '8454,8800p' $SOURCE >$DEST_BASE/02-core-specification/13-workflow.md

# Extract Part III
sed -n '8826,9651p' $SOURCE >$DEST_BASE/03-go-implementation/14-registry.md
sed -n '9653,10343p' $SOURCE >$DEST_BASE/03-go-implementation/15-validator.md
sed -n '10345,10599p' $SOURCE >$DEST_BASE/03-go-implementation/16-enricher.md
sed -n '10601,11489p' $SOURCE >$DEST_BASE/03-go-implementation/17-renderer.md
sed -n '11491,11894p' $SOURCE >$DEST_BASE/03-go-implementation/18-handlers.md
sed -n '11896,12252p' $SOURCE >$DEST_BASE/03-go-implementation/19-data-sources.md
sed -n '12254,12723p' $SOURCE >$DEST_BASE/03-go-implementation/20-testing.md

echo "Extraction complete!"
