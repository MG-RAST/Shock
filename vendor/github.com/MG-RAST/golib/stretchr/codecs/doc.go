// Provides interfaces, functions and codecs that can be used to encode and decode data to various formats.
//
// Use services/web_codec_service to easily manage and retrieve appropriate codecs for handling data in a web scenario.
//
// To write a custom codec service, simply create a type that conforms to the CodecService interface.
//
// To write a custom codec, simply create a type that conforms to the Codec interface.
//
// If you wish to customize what is encoded, also conform to the Facade interface.
// This interface allows you to provide custom data to be encoded, rather than having your object encoded directly.
package codecs
