package SHOCK::Client;

use strict;
use warnings;
no warnings('once');

use IO::Handle; # needed due to some bug in UserAgent.pm

use File::Basename;
use Data::Dumper;
use JSON;
use LWP::UserAgent;
use URI::Escape;
use HTTP::Request::Common;
use HTTP::Request::StreamingUpload;

our $global_debug = 0;

1;

sub new {
    my ($class, $shock_url, $token, $debug) = @_;
    
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
        transport_method => 'requests',
        debug => $debug || $global_debug
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
sub debug {
    my ($self) = @_;
    return $self->{debug};
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
	
	
	unless (defined $self->shock_url ) {
		die "shock_url not defined";
	}
	
	if ($self->shock_url eq '') {
		die "shock_url string empty";
	}
	
	my $my_url = $self->shock_url . "/$resource";
	
	
	
	#if (defined $self->token) {
	#	$query{'auth'}=$self->token;
	#}
	
	#build query string:
	my $query_string = "";
	
	foreach my $key (keys %query) {
		my $value = $query{$key};
		
		unless (defined $value) {
			if ((length($query_string) != 0)) {
				$query_string .= '&';
			}
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
	
	print "request: $method \"$my_url\"\n" if $self->debug;
	
	
	
	my @method_args = ($my_url, ($self->token)?('Authorization' , "OAuth ".$self->token):());
	if ($self->{'debug'} ==1) {
		if (defined $self->token) {
			print "found token: ".substr($self->token, 0, 20)."... \n";
		} else {
			print "found no token\n";
		}

	}
	
	
	my $is_download = 0;
	if (defined $headers) {
		if (defined $headers->{':content_cb'}){
			$is_download = 1;
		}
		
		push(@method_args, %$headers);
		
	}
	if ($self->{'debug'} ==1) {
		#print 'method_args: '.join(',', @method_args)."\n";
		print 'method_args: '.Dumper(@method_args)."\n";
	}
	
    #### http request
	my $response_object = undef;
	
    eval {
		
       
		if ($self->{'debug'} ==1) {
			print "invoking $method-request...\n";
		}
		
        if ($method eq 'GET') {
			$response_object = $self->agent->get(@method_args);
		} elsif ($method eq 'DELETE') {
			$response_object = $self->agent->delete(@method_args);
		} elsif ($method eq 'POST') {
			$self->agent->show_progress(1);
			$response_object = $self->agent->post(@method_args);
		} elsif ($method eq 'PUT') {
			my $request = HTTP::Request::Common::POST(@method_args); # use POST, then change to PUT in next line !
			$request->method('PUT');
	
			#if ($self->{'debug'} ==1) {
			#	print "request: ".Dumper($request)."\n";
			#}
			$response_object = $self->agent->request($request);
			#$response_object = $self->agent->put(@method_args); #does not work with multiform
		} else {
			die "method \"$method\" not implemented yet";
		}
		
		if ($self->{'debug'} ==1) {
			print "content: ".$response_object->content."\n";
		}
		if ($self->{'debug'} ==1) {
			print "response_object: ".Dumper($response_object)."\n";
		}
		
		
    };
	if ($@) {
		print STDERR "[error] ".$@."\n";
		return undef;
	}
	unless ($response_object->is_success) {
		print STDERR "response->status_line: ", $response_object->status_line, "\n";
		return undef;
	}
	
	#### JSON -> hash
	my $response_content_hash = undef;
	eval {
		$response_content_hash = $self->json->decode( $response_object->content );
	};
	
	if ($@) {
		print STDERR "[error] ".$@."\n";
		if (! ref($response_content_hash) && ($is_download==0 )) {
			print STDERR "[error] unable to connect to Shock ".$self->shock_url." response_content is not a reference\n";
		} elsif (exists($response_content_hash->{error}) && $response_content_hash->{error}) {
			print STDERR "[error] unable to send $method request to Shock: ".$response_content_hash->{error}[0]."\n";
		}
		return undef;
	}
	
	if (exists($response_content_hash->{error}) && $response_content_hash->{error}) {
		print STDERR "[error] unable to send $method request to Shock: ".$response_content_hash->{error}[0]."\n";
		return undef;
	}
	
	if ($self->{'debug'} ==1) {
		print "returning response_object...\n";
	}
	
	return $response_content_hash;
	
}

# basic requests
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

sub put {
	my ($self, $resource, $query, $headers) = @_;
	
	return $self->request('PUT', $resource, $query, $headers);
}

sub post_node {
    my ($self, $node, $query, $headers ) = @_;
    
	return $self->post('node/'.$node, $query, $headers);
}




sub delete_node {
    my ($self, $node) = @_;
    
	return $self->delete('node/'.$node);
}

# query attributes (!= node)
sub query { # https://github.com/MG-RAST/Shock/
		
	my ($self, %query) = @_;
	
	$query{'query'} = undef;
	
	my $response = $self->get('node', \%query);
	#print Dumper ($response);
	
	if (defined $response->{'error'}) {
		print STDERR Dumper ($response);
		return undef;
	}
	
	unless (defined $query{'limit'}) {

		unless (defined $response->{'total_count'}) {
			die;
		}
		
		if ($response->{'total_count'} > 25) {
			# get all nodes
			$query{'limit'} = $response->{'total_count'};
			$response = $self->get('node', \%query);
		}
	}
	
	return $response;
	
}

#query node (!= attributes)
#this allows querying of fields outside of attributes section
sub querynode { # https://github.com/MG-RAST/Shock/wiki/API
	
	my ($self, %query) = @_;
	
	$query{'querynode'} = undef;
	
	my $limit = $query{'limit'};
	
	my $response = $self->get('node', \%query);
	#print Dumper ($response);
	
	if (defined $response->{'error'}) {
		print STDERR Dumper ($response);
		return undef;
	}
	
	unless (defined $response->{'total_count'}) {
		die;
	}
	
	if (defined $limit) {
		return $response;
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
        print STDERR "[error, get_node] missing node\n";
        return undef;
    }
    
	return $self->get('/node/'.$node);
	
}


sub put_node {
    my ($self, $node, $query, $headers) = @_;
    
    unless ($node) {
        print STDERR "[error, put_node] missing node\n";
        return undef;
    }
    
	return $self->put('/node/'.$node, $query, $headers);
	
}


sub set_node_attributes {
	my ($self, $node, $attributes_json_string) = @_;
	
	
	
	
	return $self->put_node($node, undef, {Content_Type => 'multipart/form-data', Content => {"attributes_str" => $attributes_json_string}});
	#return $self->put_node($node, undef, {Content => {}, Content_Type => 'multipart/form-data'});
	#return $self->put_node($node, undef, {Content_Type => 'multipart/form-data', Content => {"attributes_str" => [undef, "n/a", Content => $attributes_json_string]}});
	
	
}

sub download_to_path2 {
	 my ($self, $node, $path) = @_;
	
	
	
	open(OUTF, ">$path") || die "Can not open file $path: $!\n";
	
	my $header =  {
		':read_size_hint' => 8192,
		':content_cb'     => sub{ my ($chunk) = @_; print OUTF $chunk;}
	};
	
	my $response = $self->get('/node/'.$node, {'download' => undef}, $header);
	
		
	close OUTF;
	
	if (defined $response) {
		return $path;
	}
	system('rm -f '.$path);
	return undef;
}

sub download_to_path {
    my ($self, $node, $path) = @_;
    
    unless ($node && $path) {
        print STDERR "[error] missing node or path\n";
        return undef;
    }
    #if ($self->transport_method eq 'shock-client') {
    #    return $self->_download_shockclient($node, $path);
    #}
    
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


sub pretty {
	my ($self, $hash) = @_;
	
	return $self->json->pretty->encode ($hash);
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



#example:     upload(data => 'hello world'), where data is scalar or reference to scalar
#example:  or upload(file => 'myworld.txt')
#example:  or upload(file => 'myworld.txt', attr => {some hash})
#example:  or upload(fh => filehandle, attr => {some hash})
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
		if (ref($hash{data}) eq 'SCALAR') {
			$content->{'upload'} = [undef, "n/a", Content => ${$hash{'data'}}];
		} else {
			$content->{'upload'} = [undef, "n/a", Content => $hash{'data'}];
		}
	}
	
	if (defined $hash{'attributes'}) { # file
		#$content->{'attributes'} = [undef, "n/a", Content => $hash{'attributes'}]
		$content->{'attributes'} = [$hash{'attributes'}];
	}
	
	if (defined $hash{'attr'}) {
		$hash{'attributes_str'} = $hash{'attr'};
	}
	if (defined $hash{'attributes_str'}) { # string
		#$content->{'attributes_str'} = [undef, "n/a", Content => $hash{'attributes_str'}]
		$content->{'attributes_str'} = $hash{'attributes_str'};
	}
	
	#if (defined $hash{'attr'}) {
		# get_handle is not good
        #$content->{attributes} = $self->_get_handle($hash{attr});
	#	$content->{'attributes'} = [undef, "n/a", Content => $hash{'attr'}]
    #}
	
	
	if (defined $hash{fh}) {
		my $response = undef;
		print STDERR "using StreamingUpload\n" if $self->debug;
		
		eval {
			my $post = HTTP::Request::StreamingUpload->new(
				POST => $self->shock_url.'/node',
				fh => $hash{fh},
				headers => HTTP::Headers->new(
					'Content_Type' => 'application/octet-stream',
					'Authorization' => 'OAuth '.$self->token
					)
			);
			print Dumper($post)."\n" if $self->debug;
			my $req = LWP::UserAgent->new->request($post);
			$response = $self->json->decode( $req->content );
		};
		if ($@ || (! ref($response))) {
			print STDERR "Unable to connect to Shock server\n";
			return undef;
		} elsif (exists($response->{error}) && $response->{error}) {
			print STDERR "Unable to POST to Shock: ".$response->{error}[0]."\n";
			return undef;
		}
	
	
		return $response;
	}
	
    $HTTP::Request::Common::DYNAMIC_FILE_UPLOAD = 1;
	
	return $self->post('node', undef, {Content_Type => 'multipart/form-data', Content => $content}); # resource, query, headers
	 
}


#upload multiple files/data with attribute "temporary" to shock
#argument is a hash reference:
#example: $files->{'object1'}->{'file'} = './mylocalfile.txt';
#example: $files->{'object2'}->{'data'} = 'this is data';
#it adds  $files->{'object1'}->{'node'} = <shock node id>
sub upload_temporary_files {
	my ($self, $job_input) = @_;

	
	#check
	foreach my $input (keys(%$job_input)) {
		my $input_h = $job_input->{$input};
		if (defined($input_h->{'file'})) {
			unless (-e $input_h->{'file'}) {
				die "file ".$input_h->{'file'}." not found, input was \"$input\"";
			}
		} elsif (defined($input_h->{'data'})) {
			
		} else {
			die "not data or file found for input \"$input\"";
		}
	}
	
	#and upload job input to shock
	foreach my $input (keys(%$job_input)) {
		print "upload input object $input\n" if $self->debug;
		my $input_h = $job_input->{$input};
		
		
		my $attr = '{"temporary" : "1"}'; # I can find them later and delete them! ;-)
		
		my $node_obj=undef;
		if (defined($input_h->{'file'})) {
			print "uploading temporary ".$input_h->{'file'}." to shock...\n" if $self->debug;
			$node_obj = $self->upload('file' => $input_h->{'file'}, 'attr' => $attr);
		} elsif (defined($input_h->{'data'})) {
			print "uploading temporary data to shock...\n" if $self->debug;
			$node_obj = $self->upload('data' => $input_h->{'data'}, 'attr' => $attr);
		} else {
			die "not data or file found for input \"$input\"";
		}
		
		unless (defined($node_obj)) {
			die "could not upload to shock server";
		}
		
		if (ref($node_obj) ne 'HASH') {
			if (ref($node_obj) eq '' ) {
				print "node_obj: ".$node_obj."\n" if $self->debug;
				if ($node_obj eq '' ) {
					print "node_obj is empty string\n" if $self->debug;
				}
			}
			die "could not upload to shock server, node_obj not a hash reference, ref=".ref($node_obj);
		}
		
		unless (defined($node_obj->{'data'})) {
			print STDERR Dumper($node_obj);
			die "no data field found";
		}
		
		
		my $node = $node_obj->{'data'}->{'id'};
		unless (defined($node)) {
			print STDERR Dumper($node_obj);
			die "no node id found";
		}
		
		#print Dumper($node_obj)."\n";
		#exit(0);
		print "new node id is $node\n" if $self->debug;
		$input_h->{'node'} = $node;
		unless (defined $input_h->{'shockhost'}) {
			$input_h->{'shockhost'} = $self->shock_url(); # might have been set earlier if shock server url are different
		}
		
	}
	print "upload_temporary_files: all temporary files uploaded.\n" if $self->debug;
	
	
	return;
}

#make node readable to the world
sub permisson_readable {
	my ($self, $nodeid) = @_;
	
	unless (defined $nodeid) {
		die "nodeid not defined";
	}
	
	my $node_accls = $self->get("node/$nodeid/acl") || return undef;
	unless ($node_accls->{'status'} == 200) {
		return undef;
	}
	
	#print Dumper($node_accls);
	
	my $node_accls_read_users = $node_accls->{'data'}->{'read'} || return undef;
	
	#print "make node world readable\n";
	if (@{$node_accls_read_users} > 0) {
		my $node_accls_delete = $self->delete('node/'.$nodeid.'/acl/read/?users='.join(',', @{$node_accls_read_users})) || return undef;
		#print Dumper($node_accls_delete);
		unless ($node_accls_delete->{'status'} == 200) {
			return undef;
		}
	}

	
	return $nodeid;
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


sub cache_find {
	my ($self, $url) = @_;
	
	#check if file is already in shock
	my $query = $self->query('cache' => '1', 'cached_url' => $url)|| die;
	
	my $data_nodes = $query->{'data'} || die;
	
	if (@{$data_nodes} == 0) {
		return undef;
	} elsif (@{$data_nodes} == 1) {
		return $data_nodes->[0];
	}
	
	die "error: multiple hits for this url";
	
}

sub cache_upload {
	my ($self, $file, $url, $targetname) = @_;
	
	#upload to SHOCK
	my $shock_json =	'{'.
	' "temporary":"1",'.
	' "cache":"1",'.
	' "cached_url":"'.$url.'",'.
	' "targetname":"'.$targetname.'"'.
	'}';
	
	print "caching file in SHOCK\n" if $self->debug;
	my $up_result = $self->upload('file' => $file, 'attr' => $shock_json) || die;
	
	unless ($up_result->{'status'} == 200) {
		die;
	}
	
	my $shock_node_id = $up_result->{'data'}->{'id'} || die "SHOCK node id not found for uploaded image";
	
	# get accls
	my $node_accls = $self->get("node/$shock_node_id/acl") || die;
	unless ($node_accls->{'status'} == 200) {
		die;
	}
	
	my $node_accls_read_users = $node_accls->{'data'}->{'read'} || die;
	
	# make node world readable
	if (@{$node_accls_read_users} > 0) {
		my $node_accls_delete = $self->delete('node/'.$shock_node_id.'/acl/read/?users='.join(',', @{$node_accls_read_users})) || die;
		unless ($node_accls_delete->{'status'} == 200) {
			die;
		}
	}
	
	return $file;
}


sub cached_download {
	my ($self, $url, $dir, $targetname, $download_cmd)=@_;
	
	# example $curl_download_cmd = "cd $dir && curl $ssl -L -o $targetname --retry 1 --retry-delay 5 \"$url\"";
	
	if (length($url) < 5) {
		die;
	}
	my $file = $dir. $targetname;
	
	my $shock_node_obj = $self->cache_find($url);
	
	
	
	if (defined($shock_node_obj)) {
		# found url, download from shock
		
		my $shock_node_id = $shock_node_obj->{'data'}->{'id'};
		
		unless (defined $shock_node_id) {
			print STDERR Dumper($shock_node_obj);
			die;
		}
		
		my $download_result = $self->download_to_path($shock_node_id, $file);
		
		return $download_result;

	} else {
		# not found, download directly and upload to shock
		
		# 1) download from url
		system($download_cmd);
		unless (-s $file ) {
			die "file $file was not downloaded!?";
		}
		
		# 2) upload to SHOCK
		return cache_upload($file, $url, $targetname);
	
		
	}
	
	#
	return undef;
}


# helper function

sub getNodeFromURL {
	my ($url) = @_;
	#print "url: $url\n";
	my ($node) = $url =~ /node\/(.*)(\?download)?/;
	#print "node: $node\n";
	
	return $node;
	
}

