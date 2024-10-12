## CT_TextBodyProperties Complex Type

### Sequence Elements
- ```prstTxWarp``` (```CT_PresetTextShape```): Preset Text Shape
- **scene3d (CT_Scene3D)**: 3D Scene Properties
- **extLst (CT_OfficeArtExtensionList)**

### EG_TextAutofit Group
- **Choice**:
  - **noAutofit (CT_TextNoAutofit)**: No AutoFit
  - **normAutofit (CT_TextNormalAutofit)**: Normal AutoFit
  - **spAutoFit (CT_TextShapeAutofit)**: Shape AutoFit

### EG_Text3D Group
- **Choice**:
  - **sp3d (CT_Shape3D)**: Apply 3D shape properties
  - **flatTx (CT_FlatText)**: No text in 3D scene

### Attributes
- **rot (ST_Angle)**: Rotation
- **spcFirstLastPara (xsdboolean)**: Paragraph Spacing
- **vertOverflow (ST_TextVertOverflowType)**: Text Vertical Overflow
- **horzOverflow (ST_TextHorzOverflowType)**: Text Horizontal Overflow
- **vert (ST_TextVerticalType)**: Vertical Text
- **wrap (ST_TextWrappingType)**: Text Wrapping Type
- **lIns (ST_Coordinate32)**: Left Inset
- **tIns (ST_Coordinate32)**: Top Inset
- **rIns (ST_Coordinate32)**: Right Inset
- **bIns (ST_Coordinate32)**: Bottom Inset
- **numCol (ST_TextColumnCount)**: Number of Columns
- **spcCol (ST_PositiveCoordinate32)**: Space Between Columns
- **rtlCol (xsdboolean)**: Columns Right-To-Left
- **fromWordArt (xsdboolean)**: From WordArt
- **anchor (ST_TextAnchoringType)**: Anchor
- **anchorCtr (xsdboolean)**: Anchor Center
- **forceAA (xsdboolean)**: Force Anti-Alias
- **upright (xsdboolean)**: Text Upright
- **compatLnSpc (xsdboolean)**: Compatible Line Spacing