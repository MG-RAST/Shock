// Provides services for working with codecs.
//
// The CodecService interface defines various functions that make it much easier to obtain a
// codec object that is appropriate for the data you wish to handle.
//
// To write a new codec service, simply confrom to the CodecService interface, and install it by
// doing:
//
//    // get a service
//    codecService := NewWebCodecService()
//
//    // make your own codec
//    myCodec := new(MyCodec)
//
//    // install the codec
//    codecService.AddCodec(myCodec)
//
package services
