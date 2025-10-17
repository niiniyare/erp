#!/bin/bash

# Schema Organization Script
# This script implements the categorization plan from SCHEMA_ORGANIZATION_STRATEGY.md

echo "🔄 Starting schema organization..."

# Ensure we're in the definitions directory
cd /data/data/com.termux/files/home/project/erp/docs/ui/Schema/definitions

# Color-related properties
echo "📁 Organizing core/color properties..."
mv core/Property.*Color*.json core/color/ 2>/dev/null || true
mv core/Property.*Opacity*.json core/color/ 2>/dev/null || true
mv core/Property.Fill*.json core/color/ 2>/dev/null || true
mv core/Property.Flood*.json core/color/ 2>/dev/null || true
mv core/Property.Stop*.json core/color/ 2>/dev/null || true
mv core/Property.Stroke*.json core/color/ 2>/dev/null || true
mv core/Property.Lighting*.json core/color/ 2>/dev/null || true
mv ColorSchema.json core/color/ 2>/dev/null || true
mv PresetColor.json core/color/ 2>/dev/null || true
mv ColorMapType.json core/color/ 2>/dev/null || true

# Typography-related properties
echo "📝 Organizing core/typography properties..."
mv core/Property.Font*.json core/typography/ 2>/dev/null || true
mv core/Property.Text*.json core/typography/ 2>/dev/null || true
mv core/Property.Letter*.json core/typography/ 2>/dev/null || true
mv core/Property.Line*.json core/typography/ 2>/dev/null || true
mv core/Property.Word*.json core/typography/ 2>/dev/null || true
mv core/Property.Writing*.json core/typography/ 2>/dev/null || true
mv core/Property.Glyph*.json core/typography/ 2>/dev/null || true
mv core/Property.Hanging*.json core/typography/ 2>/dev/null || true
mv core/Property.Hyphen*.json core/typography/ 2>/dev/null || true
mv core/Property.InitialLetter*.json core/typography/ 2>/dev/null || true
mv core/Property.TabSize*.json core/typography/ 2>/dev/null || true
mv core/Property.UnicodeBidi*.json core/typography/ 2>/dev/null || true
mv core/Property.VerticalAlign*.json core/typography/ 2>/dev/null || true
mv core/Property.WhiteSpace*.json core/typography/ 2>/dev/null || true
mv core/Property.Widows*.json core/typography/ 2>/dev/null || true
mv core/Property.Orphans*.json core/typography/ 2>/dev/null || true
mv PlainSchema.json core/typography/ 2>/dev/null || true
mv MultilineTextSchema.json core/typography/ 2>/dev/null || true
mv RemarkSchema.json core/typography/ 2>/dev/null || true
mv TplSchema.json core/typography/ 2>/dev/null || true
mv WordsSchema.json core/typography/ 2>/dev/null || true

# Layout-related properties
echo "🏗️ Organizing core/layout properties..."
mv core/Property.Display*.json core/layout/ 2>/dev/null || true
mv core/Property.Flex*.json core/layout/ 2>/dev/null || true
mv core/Property.Grid*.json core/layout/ 2>/dev/null || true
mv core/Property.Box*.json core/layout/ 2>/dev/null || true
mv core/Property.*Size*.json core/layout/ 2>/dev/null || true
mv core/Property.Width*.json core/layout/ 2>/dev/null || true
mv core/Property.Height*.json core/layout/ 2>/dev/null || true
mv core/Property.Min*.json core/layout/ 2>/dev/null || true
mv core/Property.Max*.json core/layout/ 2>/dev/null || true
mv core/Property.Overflow*.json core/layout/ 2>/dev/null || true
mv core/Property.Position*.json core/layout/ 2>/dev/null || true
mv core/Property.Top*.json core/layout/ 2>/dev/null || true
mv core/Property.Bottom*.json core/layout/ 2>/dev/null || true
mv core/Property.Left*.json core/layout/ 2>/dev/null || true
mv core/Property.Right*.json core/layout/ 2>/dev/null || true
mv core/Property.Float*.json core/layout/ 2>/dev/null || true
mv core/Property.Clear*.json core/layout/ 2>/dev/null || true
mv core/Property.ZIndex*.json core/layout/ 2>/dev/null || true
mv core/Property.Visibility*.json core/layout/ 2>/dev/null || true
mv core/Property.Clip*.json core/layout/ 2>/dev/null || true
mv core/Property.Container*.json core/layout/ 2>/dev/null || true
mv core/Property.Contain*.json core/layout/ 2>/dev/null || true
mv core/Property.Align*.json core/layout/ 2>/dev/null || true
mv core/Property.Justify*.json core/layout/ 2>/dev/null || true
mv core/Property.Place*.json core/layout/ 2>/dev/null || true
mv core/Property.Order*.json core/layout/ 2>/dev/null || true
mv core/Property.AspectRatio*.json core/layout/ 2>/dev/null || true
mv ContainerSchema.json core/layout/ 2>/dev/null || true
mv FlexSchema.json core/layout/ 2>/dev/null || true
mv GridSchema.json core/layout/ 2>/dev/null || true
mv Grid2DSchema.json core/layout/ 2>/dev/null || true
mv HBoxSchema.json core/layout/ 2>/dev/null || true
mv VBoxSchema.json core/layout/ 2>/dev/null || true
mv WrapperSchema.json core/layout/ 2>/dev/null || true

# Browser compatibility
echo "🌐 Organizing core/compatibility properties..."
mv core/Property.Moz*.json core/compatibility/ 2>/dev/null || true
mv core/Property.Ms*.json core/compatibility/ 2>/dev/null || true
mv core/Property.Webkit*.json core/compatibility/ 2>/dev/null || true

# Move utility files
echo "🔧 Organizing utility files..."
mv schema_main.json utility/ 2>/dev/null || true
mv JsonSchema.json utility/ 2>/dev/null || true
mv Globals.json utility/ 2>/dev/null || true
mv ClassName.json utility/ 2>/dev/null || true
mv SchemaClassName.json utility/ 2>/dev/null || true
mv SchemaType.json utility/ 2>/dev/null || true
mv debounceConfig.json utility/ 2>/dev/null || true
mv trackConfig.json utility/ 2>/dev/null || true
mv textPositionType.json utility/ 2>/dev/null || true
mv TooltipPosType.json utility/ 2>/dev/null || true
mv MODE_TYPE.json utility/ 2>/dev/null || true
mv TestIdBuilder.json utility/ 2>/dev/null || true

# Atomic components
echo "⚛️ Organizing components/atoms..."
mv ActionSchema.json components/atoms/ 2>/dev/null || true
mv BadgeObject.json components/atoms/ 2>/dev/null || true
mv CheckboxControlSchema.json components/atoms/ 2>/dev/null || true
mv RadioControlSchema.json components/atoms/ 2>/dev/null || true
mv SwitchControlSchema.json components/atoms/ 2>/dev/null || true
mv TextControlSchema.json components/atoms/ 2>/dev/null || true
mv InputColorControlSchema.json components/atoms/ 2>/dev/null || true
mv HiddenControlSchema.json components/atoms/ 2>/dev/null || true
mv IconSchema.json components/atoms/ 2>/dev/null || true
mv IconItemSchema.json components/atoms/ 2>/dev/null || true
mv IconCheckedSchema.json components/atoms/ 2>/dev/null || true
mv ImageSchema.json components/atoms/ 2>/dev/null || true
mv LinkSchema.json components/atoms/ 2>/dev/null || true
mv TagSchema.json components/atoms/ 2>/dev/null || true
mv DividerSchema.json components/atoms/ 2>/dev/null || true
mv SpinnerSchema.json components/atoms/ 2>/dev/null || true
mv ProgressSchema.json components/atoms/ 2>/dev/null || true
mv StatusSchema.json components/atoms/ 2>/dev/null || true
mv Option.json components/atoms/ 2>/dev/null || true
mv Options.json components/atoms/ 2>/dev/null || true
mv LabelAlign.json components/atoms/ 2>/dev/null || true

# Molecular components
echo "🧬 Organizing components/molecules..."
mv AlertSchema.json components/molecules/ 2>/dev/null || true
mv AvatarSchema.json components/molecules/ 2>/dev/null || true
mv CalendarSchema.json components/molecules/ 2>/dev/null || true
mv CardSchema.json components/molecules/ 2>/dev/null || true
mv Card2Schema.json components/molecules/ 2>/dev/null || true
mv CarouselSchema.json components/molecules/ 2>/dev/null || true
mv ChartSchema.json components/molecules/ 2>/dev/null || true
mv DateControlSchema.json components/molecules/ 2>/dev/null || true
mv DateRangeControlSchema.json components/molecules/ 2>/dev/null || true
mv DateTimeControlSchema.json components/molecules/ 2>/dev/null || true
mv DropdownButtonSchema.json components/molecules/ 2>/dev/null || true
mv DropdownButton.json components/molecules/ 2>/dev/null || true
mv EditorControlSchema.json components/molecules/ 2>/dev/null || true
mv FileControlSchema.json components/molecules/ 2>/dev/null || true
mv ImageControlSchema.json components/molecules/ 2>/dev/null || true
mv InputGroupControlSchema.json components/molecules/ 2>/dev/null || true
mv InputSignatureSchema.json components/molecules/ 2>/dev/null || true
mv LocationControlSchema.json components/molecules/ 2>/dev/null || true
mv PasswordSchema.json components/molecules/ 2>/dev/null || true
mv PickerControlSchema.json components/molecules/ 2>/dev/null || true
mv QRCodeSchema.json components/molecules/ 2>/dev/null || true
mv RangeControlSchema.json components/molecules/ 2>/dev/null || true
mv RatingControlSchema.json components/molecules/ 2>/dev/null || true
mv SearchBoxSchema.json components/molecules/ 2>/dev/null || true
mv SelectControlSchema.json components/molecules/ 2>/dev/null || true
mv TextareaControlSchema.json components/molecules/ 2>/dev/null || true
mv TimeControlSchema.json components/molecules/ 2>/dev/null || true
mv VideoSchema.json components/molecules/ 2>/dev/null || true
mv AudioSchema.json components/molecules/ 2>/dev/null || true

# Organism components  
echo "🦠 Organizing components/organisms..."
mv FormSchema.json components/organisms/ 2>/dev/null || true
mv FormControlSchema.json components/organisms/ 2>/dev/null || true
mv TableSchema.json components/organisms/ 2>/dev/null || true
mv TableSchema2.json components/organisms/ 2>/dev/null || true
mv ListSchema.json components/organisms/ 2>/dev/null || true
mv CardsSchema.json components/organisms/ 2>/dev/null || true
mv NavSchema.json components/organisms/ 2>/dev/null || true
mv TabsSchema.json components/organisms/ 2>/dev/null || true
mv DialogSchema.json components/organisms/ 2>/dev/null || true
mv DrawerSchema.json components/organisms/ 2>/dev/null || true
mv WizardSchema.json components/organisms/ 2>/dev/null || true
mv StepsSchema.json components/organisms/ 2>/dev/null || true
mv PaginationSchema.json components/organisms/ 2>/dev/null || true
mv CRUD*.json components/organisms/ 2>/dev/null || true
mv MatrixControlSchema.json components/organisms/ 2>/dev/null || true
mv ComboControlSchema.json components/organisms/ 2>/dev/null || true
mv ArrayControlSchema.json components/organisms/ 2>/dev/null || true

# Template components
echo "📄 Organizing components/templates..."
mv PageSchema.json components/templates/ 2>/dev/null || true
mv ServiceSchema.json components/templates/ 2>/dev/null || true
mv OperationSchema.json components/templates/ 2>/dev/null || true
mv MappingSchema.json components/templates/ 2>/dev/null || true
mv TasksSchema.json components/templates/ 2>/dev/null || true
mv IFrameSchema.json components/templates/ 2>/dev/null || true

# Form interactions
echo "📝 Organizing interactions/forms..."
mv ConditionBuilderControlSchema.json interactions/forms/ 2>/dev/null || true
mv FormulaControlSchema.json interactions/forms/ 2>/dev/null || true
mv DiffControlSchema.json interactions/forms/ 2>/dev/null || true
mv JSONSchemaEditorControlSchema.json interactions/forms/ 2>/dev/null || true
mv TransferControlSchema.json interactions/forms/ 2>/dev/null || true
mv TransferPickerControlSchema.json interactions/forms/ 2>/dev/null || true
mv NestedSelectControlSchema.json interactions/forms/ 2>/dev/null || true
mv ChainedSelectControlSchema.json interactions/forms/ 2>/dev/null || true
mv ListControlSchema.json interactions/forms/ 2>/dev/null || true
mv TableControlSchema.json interactions/forms/ 2>/dev/null || true
mv CheckboxesControlSchema.json interactions/forms/ 2>/dev/null || true
mv RadiosControlSchema.json interactions/forms/ 2>/dev/null || true
mv ButtonGroupControlSchema.json interactions/forms/ 2>/dev/null || true

# Navigation interactions
echo "🧭 Organizing interactions/navigation..."
mv AnchorNavSchema.json interactions/navigation/ 2>/dev/null || true
mv AnchorNavSectionSchema.json interactions/navigation/ 2>/dev/null || true
mv NavItemSchema.json interactions/navigation/ 2>/dev/null || true
mv NavOverflow.json interactions/navigation/ 2>/dev/null || true
mv LinkActionSchema.json interactions/navigation/ 2>/dev/null || true
mv UrlActionSchema.json interactions/navigation/ 2>/dev/null || true
mv DialogActionSchema.json interactions/navigation/ 2>/dev/null || true
mv DrawerActionSchema.json interactions/navigation/ 2>/dev/null || true
mv ToastActionSchema.json interactions/navigation/ 2>/dev/null || true
mv CopyActionSchema.json interactions/navigation/ 2>/dev/null || true
mv EmailActionSchema.json interactions/navigation/ 2>/dev/null || true
mv ReloadActionSchema.json interactions/navigation/ 2>/dev/null || true
mv AjaxActionSchema.json interactions/navigation/ 2>/dev/null || true
mv OtherActionSchema.json interactions/navigation/ 2>/dev/null || true

# Data interactions
echo "📊 Organizing interactions/data..."
mv DataProvider.json interactions/data/ 2>/dev/null || true
mv DataProviderCollection.json interactions/data/ 2>/dev/null || true
mv ComposedDataProvider.json interactions/data/ 2>/dev/null || true
mv SchemaApi.json interactions/data/ 2>/dev/null || true
mv SchemaApiObject.json interactions/data/ 2>/dev/null || true
mv BaseApiObject.json interactions/data/ 2>/dev/null || true
mv RowSelectionSchema.json interactions/data/ 2>/dev/null || true
mv RowSelectionOptionsSchema.json interactions/data/ 2>/dev/null || true
mv ImagesSchema.json interactions/data/ 2>/dev/null || true
mv TimelineSchema.json interactions/data/ 2>/dev/null || true
mv TimelineItemSchema.json interactions/data/ 2>/dev/null || true
mv ListItemSchema.json interactions/data/ 2>/dev/null || true
mv SparkLineSchema.json interactions/data/ 2>/dev/null || true

echo "✅ Schema organization complete!"
echo "📊 Summary of organization:"
echo "   - core/: $(find core/ -name "*.json" | wc -l) files"
echo "   - components/: $(find components/ -name "*.json" | wc -l) files" 
echo "   - interactions/: $(find interactions/ -name "*.json" | wc -l) files"
echo "   - utility/: $(find utility/ -name "*.json" | wc -l) files"
echo "   - remaining: $(find . -maxdepth 1 -name "*.json" | wc -l) files"