package Shock;

use strict;
use warnings;
no warnings('once');

use File::Basename;
use Data::Dumper;
use JSON;
use LWP::UserAgent;
use URI::Escape;

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

sub create_url {
	my ($self, $resource, %query) = @_;
	
	my $my_url = $self->shock_url . "/$resource";
	
	#if (defined $self->token) {
	#	$query{'auth'}=$self->token;
	#}
	
	#build query string:
	my $query_string = "";
	
	foreach my $key (keys %query) {
		my $value = $query{$key};
		
		unless (defined $value) {
			$query_string .= $key;
			next;
		}
		
		my @values=();
		if (ref($value) eq 'ARRAY') {
			@values=@$value;
		} else {
			@values=($value);
		}
		
		foreach my $value (@values) {
			if ((length($query_string) != 0)) {
				$query_string .= '&';
			}
			$query_string .= $key.'='.$value;
		}
	}
	
	
	if (length($query_string) != 0) {
		
		#print "url: ".$my_url.'?'.$query_string."\n";
		$my_url .= '?'.$query_string;#uri_escape()
	}
	
	
	
	
	return $my_url;
}


sub request {
	#print 'request: '.join(',',@_)."\n";
	my ($self, $method, $resource, $query, $headers) = @_;
	
	
	my $my_url = $self->create_url($resource, (defined($query)?%$query:()));
	
	print "url: $my_url\n";
	
	
	
	my @method_args = ($my_url, ($self->token)?('Authorization' , "OAuth ".$self->token):());
	
	if (defined $headers) {
		push(@method_args, %$headers);
	}
	
	#print 'method_args: '.join(',', @method_args)."\n";
	
	my $response_content = undef;
    
    eval {
		
        my $response_object = undef;
		
        if ($method eq 'GET') {
			$response_object = $self->agent->get(@method_args );
		} elsif ($method eq 'DELETE') {
			$response_object = $self->agent->delete(@method_args );
		} elsif ($method eq 'POST') {
			$response_object = $self->agent->post(@method_args );
		} else {
			die "not implemented yet";
		}

		
		$response_content = $self->json->decode( $response_object->content );
        
    };
    
	if ($@ || (! ref($response_content))) {
        print STDERR "[error] unable to connect to Shock ".$self->shock_url."\n";
        return undef;
    } elsif (exists($response_content->{error}) && $response_content->{error}) {
        print STDERR "[error] unable to send $method request to Shock: ".$response_content->{error}[0]."\n";
		return undef;
    } else {
        return $response_content;
    }
	
}


sub get {
	#print 'get: '.join(',',@_)."\n";
	my ($self, $resource, $query, $headers) = @_;
	
	return $self->request('GET',  $resource, $query, $headers);
}

sub delete {
	my ($self, $resource, $query, $headers) = @_;
	
	return $self->request('DELETE', $resource, $query, $headers);
}

sub post {
	#print 'get: '.join(',',@_)."\n";
	my ($self, $resource, $query, $headers) = @_;
	
	return $self->request('POST', $resource, $query, $headers);
}

sub delete_node {
    my ($self, $node) = @_;
    
	return $self->delete('node/'.$node);
}


sub query { # https://github.com/MG-RAST/Shock#get_nodes
		
	my ($self, %query) = @_;
	
	$query{'query'} = undef;
	
	my $response = $self->get('node', \%query);
	#print Dumper ($response);
	unless (defined $response->{'total_count'}) {
		die;
	}
	
	if ($response->{'total_count'} > 25) {
		# get all nodes
		$query{'limit'} = $response->{'total_count'};
		$response = $self->get('node', \%query);
	}
	
	
	return $response;
	
}


#get('/node/'.$node, %h);
sub get_node {
    my ($self, $node) = @_;
    
    unless ($node) {
        print STDERR "[error] missing node\n";
        return undef;
    }
    
	return $self->get('/node/'.$node);
	
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
		
		my @auth = ($self->token)?('Authorization' , "OAuth ".$self->token):();
		
        
        $get = $self->agent->get( $self->shock_url.'/node/'.$node.'?download',
                                      @auth,
                                      ':read_size_hint' => 8192,
                                      ':content_cb'     => sub{ my ($chunk) = @_; print OUTF $chunk; } );
        close OUTF;
        $content = $get->content;
    };
    
    if ($@) {
        print STDERR "[error] unable to connect to Shock ".$self->shock_url."\n";
		unlink($path);
        return undef;
    } elsif (ref($content) && exists($content->{error}) && $content->{error}) {
        print STDERR "[error] unable to GET file $node from Shock: ".$content->{error}[0]."\n";
		unlink($path);
        return undef;
    } elsif (! -s $path) {
        print STDERR "[error] unable to download to $path: $!\n";
		unlink($path);
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

#example:     upload(data => 'hello world')
#example:  or upload(file => 'myworld.txt')
#example:  or upload(file => 'myworld.txt', attr => {some hash})
# TODO implement PUT here or in another function
sub upload {
    my ($self, %hash) = @_;
	
    #my $response = undef;
    my $content = {};
    #my $url = $self->shock_url.'/node';
    #my $method = 'POST';
    #if ($hash{'node'}) {
    #    $url = $url.'/'.$hash{'node'};
    #    $method = 'PUT';
    #}
	
	if (defined $hash{file}) {
		unless (-s $hash{file}) {
			die "file not found".$hash{'file'};
		}
		$content->{'upload'} = [$hash{'file'}]
	}
	if (defined $hash{data}) {
		$content->{'upload'} = [undef, "n/a", Content => $hash{'data'}]
	}
	
   
	if (defined $hash{'attr'}) {
		# get_handle is not good
        #$content->{attributes} = $self->_get_handle($hash{attr});
		$content->{'attributes'} = [undef, "n/a", Content => $hash{'attr'}]
    }
    
    $HTTP::Request::Common::DYNAMIC_FILE_UPLOAD = 1;
	
	return $self->post('node', undef, {Content_Type => 'multipart/form-data', Content => $content});
	
#	
#    eval {
#        my $res = undef;
#		my @auth = ($self->token)?('Authorization' , "OAuth ".$self->token):();
#		
#        if ($method eq 'POST') {
#			$res = $self->agent->post($url, Content_Type => 'multipart/form-data', @auth, Content => $content);
#		} else {
#			$res = $self->agent->put($url, Content_Type => 'multipart/form-data', @auth, Content => $content);
#        }
#        $response = $self->json->decode( $res->content );
#    };
#    if ($@ || (! ref($response))) {
#        print STDERR "[error] unable to connect to Shock ".$self->shock_url."\n";
#        return undef;
#    } elsif (exists($response->{error}) && $response->{error}) {
#        print STDERR "[error] unable to $method data to Shock: ".$response->{error}[0]."\n";
#    } else {
#        return $response->{data};
#    }
}


#upload multiple files/data with attribute "temporary" to shock
#argument is a hash reference:
#example: $files->{'object1'}->{'file'} = './mylocalfile.txt';
#example: $files->{'object2'}->{'data'} = 'this is data';
#it adds  $files->{'object1'}->{'node'} = <shock node id>
sub upload_temporary_files {
	my ($self, $job_input) = @_;

	
	#and upload job input to shock
	foreach my $input (keys(%$job_input)) {
		my $input_h = $job_input->{$input};
		
		
		my $attr = '{"temporary" : "1"}'; # I can find them later and delete them! ;-)
		
		my $node_obj=undef;
		if (defined($input_h->{'file'})) {
			print "uploading temporary ".$input_h->{'file'}." to shock...\n";
			$node_obj = $self->upload('file' => $input_h->{'file'}, 'attr' => $attr);
			print "uploaded.\n";
		} elsif (defined($input_h->{'data'})) {
			print "uploading temporary data to shock...\n";
			$node_obj = $self->upload('data' => $input_h->{'data'}, 'attr' => $attr);
			print "uploaded.\n";
		} else {
			die "not data or file found";
		}
		
		unless (defined($node_obj)) {
			die "could not upload to shock server";
		}
		my $node = $node_obj->{'data'}->{'id'};
		unless (defined($node)) {
			print Dumper($node_obj);
			die;
		}
		
		#print Dumper($node_obj)."\n";
		#exit(0);
		print "new node is $node\n";
		$input_h->{'node'} = $node;
		$input_h->{'shockhost'} = $self->shock_url();
		
	}
	print "upload_temporary_files: all temporary files uploaded.\n";
	
	
	return;
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
    
	eval {
		if (-s $item) {
			return [$item];
		}
	};
	# TODO: this is ugly.
	
	return [undef, "n/a", Content => $item];
}






