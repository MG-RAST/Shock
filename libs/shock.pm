package Shock;

use strict;
use warnings;
no warnings('once');

use File::Basename;
use Data::Dumper;
use JSON;
use LWP::UserAgent;

1;

sub new {
    my ($class, $shock_url, $token) = @_;
    
    my $agent = LWP::UserAgent->new;
    my $json = JSON->new;
    $json = $json->utf8();
    $json->max_size(0);
    $json->allow_nonref;
    
    my $self = {
        json => $json,
        agent => $agent,
        shock_url => $shock_url || '',
        token => $token || '',
        transport_method => 'requests'
    };
    if (system("type shock-client > /dev/null 2>&1") == 0) {
        $self->{transport_method} = 'shock-client';
    }

    bless $self, $class;
    return $self;
}

sub json {
    my ($self) = @_;
    return $self->{json};
}
sub agent {
    my ($self) = @_;
    return $self->{agent};
}
sub shock_url {
    my ($self) = @_;
    return $self->{shock_url};
}
sub token {
    my ($self) = @_;
    return $self->{token};
}
sub transport_method {
    my ($self) = @_;
    return $self->{transport_method};
}

sub _set_shockclient_auth {
    my ($self) = @_;
    
    if ($self->token && ($self->transport_method eq 'shock-client')) {
        my $auth = $self->json->encode( {"access_token" => $self->token} );
        my $msg = `shock-client auth set-token $auth`;
        if (($? >> 8) == 0) {
            return 1;
        } else {
            print STDERR "[error] setting auth token in shock-client: $msg\n";
            return 0;
        }
    } else {
        print STDERR "[error] missing token or shock-client\n";
        return 0;
    }
}

sub get_node {
    my ($self, $node) = @_;
    
    unless ($node) {
        print STDERR "[error] missing node\n";
        return undef;
    }
    
    my $content = undef;    
    eval {
        my $get = undef;
        if ($self->token) {
            $get = $self->agent->get($self->shock_url.'/node/'.$node, 'Authorization' => "OAuth ".$self->token);
        } else {
            $get = $self->agent->get($self->shock_url.'/node/'.$node);
        }
        $content = $self->json->decode($get->content);
    };
    
    if ($@ || (! ref($content))) {
        print STDERR "[error] unable to connect to Shock ".$self->shock_url."\n";
        return undef;
    } elsif (exists($content->{error}) && $content->{error}) {
        print STDERR "[error] unable to GET node $node from Shock: ".$content->{error}[0]."\n";
        return undef;
    } else {
        return $content->{data};
    }
}

sub download_to_path {
    my ($self, $node, $path) = @_;
    
    unless ($node && $path) {
        print STDERR "[error] missing node or path\n";
        return undef;
    }
    if ($self->transport_method eq 'shock-client') {
        return $self->_download_shockclient($node, $path);
    }
    
    my $content = undef;
    eval {
        my $get = undef;
        open(OUTF, ">$path") || die "Can not open file $path: $!\n";
        if ($self->token) {
            $get = $self->agent->get( $self->shock_url.'/node/'.$node.'?download',
                                      'Authorization'   => "OAuth ".$self->token,
                                      ':read_size_hint' => 8192,
                                      ':content_cb'     => sub{ my ($chunk) = @_; print OUTF $chunk; } );
        } else {
            $get = $self->agent->get( $self->shock_url.'/node/'.$node.'?download',
                                      ':read_size_hint' => 8192,
                                      ':content_cb'     => sub{ my ($chunk) = @_; print OUTF $chunk; } );
        }
        close OUTF;
        $content = $get->content;
    };
    
    if ($@) {
        print STDERR "[error] unable to connect to Shock ".$self->shock_url."\n";
        return undef;
    } elsif (ref($content) && exists($content->{error}) && $content->{error}) {
        print STDERR "[error] unable to GET file $node from Shock: ".$content->{error}[0]."\n";
        return undef;
    } elsif (! -s $path) {
        print STDERR "[error] unable to download to $path: $!\n";
        return undef;
    } else {
        return $path;
    }
}

sub _download_shockclient {
    my ($self, $node, $path) = @_;
    
    unless ($self->_set_shockclient_auth()) {
        return undef;
    }
    my $msg = `shock-client pdownload -threads=4 $node $path`;
    if (($? >> 8) != 0) {
        print STDERR "[error] unable to download via shock-client: $node => $path: $msg\n";
        return undef;
    }
    return $path;
}

sub create_node {
    my ($self, $data, $attr) = @_;
    return $self->upload(undef, $data, $attr);
}

sub upload {
    my ($self, $node, $data, $attr) = @_;
    
    if (($self->transport_method eq 'shock-client') && (! $node) && (-s $data)) {
        my $res = $self->_upload_shockclient($data);
        if (! $attr) {
            return $res;
        } else {
            $node = $res->{id};
            $data = undef;
        }
    }
    
    my $response = undef;
    my $content = {};
    my $url = $self->shock_url.'/node';
    my $method = 'POST';
    if ($node) {
        $url = $url.'/'.$node;
        $method = 'PUT';
    }
    if ($data) {
        $content->{upload} = $self->_get_handle($data);
    }
	if ($attr) {
        $content->{attributes} = $self->_get_handle($attr);
    }
    
    $HTTP::Request::Common::DYNAMIC_FILE_UPLOAD = 1;
    eval {
        my $res = undef;
        if ($method eq 'POST') {
            if ($self->token) {
                $res = $self->agent->post($url, Content_Type => 'multipart/form-data', Authorization => "OAuth ".$self->token, Content => $content);
            } else {
                $res = $self->agent->post($url, Content_Type => 'multipart/form-data', Content => $content);
            }
        } else {
            if ($self->token) {
                $res = $self->agent->put($url, Content_Type => 'multipart/form-data', Authorization => "OAuth ".$self->token, Content => $content);
            } else {
                $res = $self->agent->put($url, Content_Type => 'multipart/form-data', Content => $content);
            }
        }
        $response = $self->json->decode( $res->content );
    };
    if ($@ || (! ref($response))) {
        print STDERR "[error] unable to connect to Shock ".$self->shock_url."\n";
        return undef;
    } elsif (exists($response->{error}) && $response->{error}) {
        print STDERR "[error] unable to $method data to Shock: ".$response->{error}[0]."\n";
    } else {
        return $response->{data};
    }
}

sub _upload_shockclient {
    my ($self, $path) = @_;
    
    unless ($self->_set_shockclient_auth()) {
        return undef;
    }
    my $msg = `shock-client pcreate -threads=4 -full $path`;
    if (($? >> 8) != 0) {
        print STDERR "[error] unable to upload via shock-client: $path: $msg\n";
        return undef;
    }
    my $res = '';
    foreach my $line (split(/\n/, $msg)) {
        chomp $line;
        if ($line !~ /Uploading/) {
            $res .= $line;
        }
    }
    return $self->json->decode($res);
}

sub _get_handle {
    my ($self, $item) = @_;
    
    if (-s $item) {
        return [$item];
    } else {
        return [undef, "n/a", Content => $item];
    }
}


