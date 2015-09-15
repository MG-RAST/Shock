// A codec for handling simple XML encoding and decoding.
//
// Simple XML is a subset of XML that keeps the data descriptions simple, yet as
// powerful and flexible as JSON.
//
// Simple XML
//
// A single object looks like this:
//
//     <object>
//       <field1>value</field1>
//       <field2>value</field2>
//       <field3>value</field3>
//     </object>
//
// Where the "object" is literal, and the "field*" elements would be your field
// names.
//
// A collection of objects looks like this:
//
//     <objects>
//       <object>
//         <field1>value</field1>
//         <field2>value</field2>
//         <field3>value</field3>
//       </object>
//       <object>
//         <field1>value</field1>
//         <field2>value</field2>
//         <field3>value</field3>
//       </object>
//       <object>
//         <field1>value</field1>
//         <field2>value</field2>
//         <field3>value</field3>
//       </object>
//     </objects>
//
// Where the "objects" and "object" tags are literal, and the "field*" elements
// would be your field names.
//
// Simple XML supports nesting objects inside each other:
//
//     <object>
//       <field1>value</field1>
//       <field2>
//           <subfield1>value</subfield1>
//           <subfield2>value</subfield2>
//           <subfield3>value</subfield3>
//       </field2>
//     </object>
//
// All values are treated as strings unless a 'type' attribute is applied to the field.
// Acceptable types values are:
//
// - int      (integer)
// - uint     (unsigned integer)
// - float    (floating point number)
// - bool     (boolean; true or false)
// - string   (string - the default)
package xml
