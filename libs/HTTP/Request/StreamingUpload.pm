package HTTP::Request::StreamingUpload;
use strict;
use warnings;
our $VERSION = '0.01';

use Carp ();
use HTTP::Request;

sub new {
    my($class, $method, $uri, %args) = @_;

    my $headers = $args{headers};
    if ($headers) {
        if (ref $headers eq 'HASH') {
            $headers = +[ %{ $headers } ];
        }
    }

    my $req = HTTP::Request->new($method, $uri, $headers);
    _set_content($req, \%args);
    $req;
}

sub _set_content {
    my($req, $args) = @_;

    if ($args->{content}) {
        $req->content($args->{content});
    } elsif ($args->{callback} && ref($args->{callback}) eq 'CODE') {
        $req->content($args->{callback});
    } elsif ($args->{path} || $args->{fh}) {
        my $fh;
        if ($args->{fh}) {
            $fh = $args->{fh};
        } else {
            open $fh, '<', $args->{path} or Carp::croak "$args->{path}: $!";
        }
        my $chunk_size = $args->{chunk_size} || 4096;
        $req->content(sub {
            my $len = read($fh, my $buf, $chunk_size);
            return unless $len;
            return $buf;
        });
    }
}

# some code takes by LWP::Protocol::http->request
sub slurp {
    my(undef, $req) = @_;
    my $content_ref = $req->content_ref;
    $content_ref = ${ $content_ref } if ref ${ $content_ref };

    my $content;
    if (ref($content_ref) eq 'CODE') {
        while (1) {
            my $buf = $content_ref->();
            last unless defined $buf;
            $content .= $buf;
        }
    } else {
        $content = ${ $content_ref };
    }
    $content;
}

1;
__END__

=for stopwords filepath callback chunked HeaderName HeaderValue fh

=head1 NAME

HTTP::Request::StreamingUpload - streaming upload wrapper for HTTP::Request

=head1 SYNOPSIS

=head2 upload from filepath

  my $req = HTTP::Request::StreamingUpload->new(
      PUT     => 'http://example.com/foo.cgi',
      path    => '/your/upload.jpg',
      headers => HTTP::Headers->new(
          'Content-Type'   => 'image/jpeg',
          'Content-Length' => -s '/your/upload.jpg',
      ),
  );
  my $res = LWP::UserAgent->new->request($req);

=head2 upload from filehandle

  open my $fh, '<', '/your/upload/requestbody' or die $!;
  my $req = HTTP::Request::StreamingUpload->new(
      PUT     => 'http://example.com/foo.cgi',
      fh      => $fh,
      headers => HTTP::Headers->new(
          'Content-Length' => -s $fh,
      ),
  );
  my $res = LWP::UserAgent->new->request($req);

=head2 upload from callback

  my @chunk = qw( foo bar baz );
  my $req = HTTP::Request::StreamingUpload->new(
      PUT      => 'http://example.com/foo.cgi',
      callback => sub { shift @chunk },
      headers => HTTP::Headers->new(
          'Content-Type'   => 'text/plain',
          'Content-Length' => 9,
      ),
  );
  my $res = LWP::UserAgent->new->request($req);

=head1 DESCRIPTION

HTTP::Request::StreamingUpload is streaming upload wrapper for L<HTTP::Request>.
It could be alike when $DYNAMIC_FILE_UPLOAD of L<HTTP::Request::Common> was used.
However, it is works only for POST method with form-data.
HTTP::Request::StreamingUpload works on the all HTTP methods.

Of course, you can big file upload using few memory by this wrapper.

=head1 HTTP::Request::StreamingUpload->new( $method, $uir, %args );

=head2 %args Options

=over 4

=item headers => [ HeaderName => 'HeaderValue', ... ]

=item headers => { HeaderName => 'HeaderValue', ... }

=item headers => HTTP::Headers->new(  HeaderName => 'HeaderValue', ... )

header is passed. HASHREF, ARRAYREF or L<HTTP::Headers> object can be passed.

If you are possible, you should set up C<Content-Length> for file upload.
However, chunked upload for HTTP 1.1 will be performed by L<LWP::UserAgent> if it does not set up.

=item path => '/your/file.txt'

set the upload file path.

=item fh => $fh

set the file-handle of upload file.
It can use instead of C<path>.

=item chunk_size => 4096

set the buffer size when reading the file of C<fh> or C<path>.

=item callback => sub { ...; return if $eof; return $buf }

Instead of C<path> or C<fh>, upload data is controlled by itself and can be made.

    # 10 times send epoch time
    callback => sub {
        return if $i++ > 10;
        return time() . "\n";
    },

=back

=head1 AUTHOR

Kazuhiro Osawa E<lt>yappo <at> shibuya <döt> plE<gt>

=head1 SEE ALSO

L<HTTP::Request>,
L<HTTP::Request::Common>,
L<HTTP::Headers>,
L<LWP::UserAgent>

=head1 LICENSE

This library is free software; you can redistribute it and/or modify
it under the same terms as Perl itself.

=cut
