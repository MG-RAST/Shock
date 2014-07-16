#!/usr/bin/env perl

use strict;
use warnings;

use lib ".";
use SHOCK::Client;

eval "use USAGEPOD qw(parse_options); 1"
or die "module USAGEPOD.pm required: wget https://raw.github.com/MG-RAST/MG-RAST-Tools/master/tools/lib/USAGEPOD.pm";


my $shockurl = $ENV{'SHOCK_SERVER_URL'} || '';
my $shocktoken = $ENV{'GLOBUSONLINE'} || $ENV{'KB_AUTH_TOKEN'};

my ($h, $help_text) = &parse_options (
'name' => 'shockclient.pl',
'version' => '1',
'synopsis' => 'shockclient.pl --show <nodeid or filename>',
'examples' => 'shockclient.pl --upload *.fasta',
'authors' => 'Wolfgang Gerlach',
'options' => [
	'',
	'Actions:',
	[ 'show'						, "show shock node and its attributes"],
	[ 'upload'						, "upload files to Shock, file as parameter"],
	[ 'delete'						, "delete Shock node"],
	[ 'query=s'						, "querystring, e.g. key=value,key2=value2"],
	[ 'querynode=s'					, "querystring, e.g. key=value,key2=value2, this allows querying of fields outside of attributes section"],
	[ 'download'					, ""],
	[ 'modify_attr=s'				, "modify nested attributes by json diff '{\"test\":\"hello\", \"name\":\"\"}', empty strings do delete (arrays not well suppoeted yet!)"],
	[ 'makepublic'					, "make node public"],
	[ 'clean_tmp'					, ""],
	'',
	'Options:',
#	[ 'xx=s'						, "xx"],
	[ 'public'						, "uploaded files will be public (default private)"],
	[ 'attributes_string=s'         , "string containing attributes to be uploaded with file, must be valid JSON" ],
	[ 'attributes_file=s'           , "file containing attributes to be uploaded with file, must be valid JSON" ],
	[ 'preview'						, "(modify_attr only): do make changes, just show result"],
	[ 'url=s' 						, "url to Shock server (default $shockurl)" ],
	[ 'token=s' 					, "default from \$KB_AUTH_TOKEN" ],
	[ 'ids'     	                , "return only node ids, formatted as comma-separated values" ],
	[ 'id_list'                     , "return only node ids, one id per line" ],
	[ 'debug' 					    , "more verbose mode for debugging (default off)" ],
	[ 'help|h'						, "", { hidden => 1  }]
	]
);



if ($h->{'help'} || keys(%$h)==0) {
	print $help_text;
	exit(0);
}

if ($h->{'url'}) {
	$shockurl = $h->{'url'};
}

if ($h->{'token'}) {
	$shocktoken = $h->{'token'};
}

my $debug = defined($h->{"debug"}) ? 1 : 0;

print "connect to SHOCK\n" if $debug;
my $shock = new SHOCK::Client($shockurl, $shocktoken, $debug); # shock production
unless (defined $shock) {
	die "error creating shock handle";
}



sub merge_hash {
	my ($attr, $changes_hash) = @_;
	
	if (ref($attr) ne ref($changes_hash) ) {
		die "hash types do not match, ".ref($attr)." and ".ref($changes_hash);
	}
	
	if (ref($attr) eq 'HASH' ) {
		foreach my $key (keys %$changes_hash) {
			my $value = $changes_hash->{$key};
			if (ref($value) eq '') {
				if ($value ne '') {
					$attr->{$key} = $value;
				} else {
					delete $attr->{$key};
				}
			} elsif (ref($value) eq 'HASH') {
				# recurse
				unless (defined $attr->{$key}) {
					$attr->{$key} = {};
				}
				merge_hash($attr->{$key}, $value);
			} elsif (ref($value) eq 'ARRAY') {
				# recurse
				unless (defined $attr->{$key}) {
					$attr->{$key} = [];
				}
				merge_hash($attr->{$key}, $value);
			} else {
				die "got $key : ".ref($value);
			}
			
		}
		
	} elsif (ref($attr) eq 'ARRAY' ) {
		
		#die "array not supported yet, too complicated.. ;-)"
		#foreach $key (keys %$changes_hash) {
		for (my $i=0 ; $i  < @{$changes_hash}; ++$i) {
			
			my $value = $changes_hash->[$i];
			my $old_value = $attr->[$i];
			
			if (ref($value) ne ref($old_value)) {
				die;
			}
			
			if (ref($value) eq '') {
				$attr->[$i] = $value;
			} elsif (ref($value) eq 'HASH') {
				# recurse
				merge_hash($attr->[$i], $value);
			} elsif (ref($value) eq 'ARRAY') {
				# recurse
				merge_hash($attr->[$i], $value);
			} else {
				die "$value of type ".ref($value);
			}
			
			
		}
		
	} else {
		die;
	}
	
}

# print node IDs or complete node structs
# 1. arg is array ref of nodes
# 2. arg is array ref of string IDs
sub print_nodes {
	my ($nodes, $node_ids) = shift(@_);
	
	
	
	if (defined $nodes) {
		
		if (defined $node_ids) {
			die "specify only nodes or node_ids";
		}
		
		$node_ids = \map {$_->{'id'}} @{$nodes};
	}
	
	unless (defined $node_ids) {
		die "no nodes found";
	}
	
	
	if (defined($h->{"id_list"})) {
		print join("\n", @{$node_ids})."\n";
	} elsif (defined($h->{"ids"})) {
		print join(',', @{$node_ids})."\n";
	} else {
		unless (defined $nodes) {
			die "node objects not found";
		}
	    pprint_json($nodes);
	}
		
	return;
	
}




my $value = undef;

if (defined($value = $h->{"query"})) {
	
	my @queries = split(/,|\=/, $value);
	my $response =  $shock->query(@queries);
	
	print_nodes($response->{'data'}, undef);
		
	exit(0);
} elsif (defined($value=$h->{"modify_attr"})) {

	
	
	my @nodes = split(',', join(',', @ARGV)); # get comma and space separated nodes
	
	print 'converting json into hash: '.$value."\n";
	my $changes_hash = $shock->json->decode( $value );
	pprint_json($changes_hash);
		
	
	foreach my $node (@nodes) {
		my $response =  $shock->get_node($node);
		pprint_json($response);
		
		my $attr = $response->{'data'}->{'attributes'};
		print "original attributes:\n";
		pprint_json($attr);
		
		merge_hash($attr, $changes_hash);
		print "modified attributes:\n";
		pprint_json($attr);
		
		my $new_attributes_string = $shock->json->pretty(0)->encode($attr);
		print "new_attributes_string: $new_attributes_string\n";
		#$shock->{'debug'} = 1;
		unless (defined $h->{"preview"}) {
			
			
			#my $ret = $shock->put_node($node, undef, {"attributes_str" => $new_attributes_string});
			my $ret = $shock->set_node_attributes($node, $new_attributes_string);
			if (defined($ret)) {
				if (defined $ret->{'error'}) {
					print STDERR "response error: ". $ret->{'error'}->[0]."\n";
				}
			}
		}
		
	}
	
	exit(0);
	
} elsif (defined($value = $h->{"querynode"})) {
	#this allows querying of fields outside of attributes section
	my @queries = split(/,|\=/, $value);
	my $response =  $shock->querynode(@queries);
	
	print_nodes($response->{'data'}, undef);
	
	exit(0);
} elsif (defined($h->{"delete"})) {
	
	
	my @nodes = split(',', join(',', @ARGV)); # get comma and space separated nodes
	
	
	foreach my $node (@nodes) {
		my $response =  $shock->delete_node($node);
		pprint_json($response);
	}
	
	
	
	exit(0);
} elsif (defined($h->{"upload"})) {
	
	
	my @files = @ARGV;
	my $attr_value = undef;
	my $attr_type = "useless_dummy";
	if (defined($h->{"attributes_string"})) {
	    #$attr = $shock->json->decode($h->{"attributes_string"});
		$attr_type = "attributes_str";
		$attr_value = $h->{"attributes_string"};
	} elsif (defined($h->{"attributes_file"}) ) {
		$attr_type = "attributes";
		$attr_value = $h->{"attributes_file"};
		unless (-s $attr_value) {
			die "file $attr_value not found";
		}
		
	    #open(JSON, $h->{"attributes_file"}) or die $!;
        #my $json_str = do { local $/; <JSON> };
        #close(JSON);
        #$attr = $shock->json->decode($json_str);
	}
	
	my $uploaded_nodes=[];
	
	foreach my $file (@files) {
		
		print "uploading ".$file."...\n" if $debug;
		
		#my $shock_node = $attr ? $shock->upload('file' => $file, 'attr' => $attr) : $shock->upload('file' => $file);
		my $shock_node = $shock->upload('file' => $file, $attr_type => $attr_value);
		unless (defined $shock_node) {
			die "unknown error uploading $file";
		}
		
		if (defined $shock_node->{'error'}) {
			die $shock_node->{'error'};
		}
		
		my $id = $shock_node->{'data'}->{'id'};
		unless (defined $id) {
				pprint_json($shock_node) if $debug;
				die "id not found";
		}
		print $file." saved with id $id\n";
		
		push (@{$uploaded_nodes}, $shock_node->{'data'});
		
		if (defined $h->{"public"}) {
			print "make id $id public...\n" if $debug;
			$shock->permisson_readable($id);
		}
	}
	
	print_nodes($uploaded_nodes, undef);
	
	exit(0);
} elsif (defined($h->{"makepublic"})) {
	my @nodes = split(',', join(',', @ARGV)); # get comma and space separated nodes
	foreach my $node (@nodes) {
		print "make id $value public...\n" if $debug;
		$shock->permisson_readable($value);
	}
	exit(0);
} elsif (defined($h->{"show"})) {
	
	
	my @nodes = split(',', join(',', @ARGV)); # get comma and space separated nodes
	
	
	foreach my $node (@nodes) {
		my $response =  $shock->get_node($node);
		#push (@{$view_nodes}, $shock_node->{'data'});
		#pprint_json($response);
		print_nodes($response->{'data'}, undef);
	}
	
	
	
	exit(0);
} elsif (defined($h->{"download"})) {
	
	
	my @nodes = split(',', join(',', @ARGV)); # get comma and space separated nodes
	
	
	foreach my $node (@nodes) {
		
		my $view_response =  $shock->get_node($node);
		pprint_json($view_response) if $debug;
		#exit(0);
		
		my $filename  = $view_response->{'data'}->{'file'}->{'name'};
		unless (defined $filename) {
			die "filename not defined, cannot save.";
		}
		
		if (-e $filename) {
			die "file \"$filename\" already exists";
		}
		
		my $response = $shock->download_to_path($node, $filename);
		if (! $response) {
		    die "error downloading $node";
		} else {
			if (-e $filename) {
				print "Node $node downloaded to $filename\n";
			} else {
				die "file \"$filename\" not found";
			}
		}
		
		
	}
	
	
	
	exit(0);
	
} elsif (defined($h->{"clean_tmp"})) {
	
	#my $shock = new SHOCK::Client($shockurl, $shocktoken);
	#unless (defined $shock) {
	#	die;
	#}
	
	my $response =  $shock->query('temporary' => 1);
	
	#my $response =  $shock->query('statistics.length_max' => 1175);
	pprint_json($response) if $debug;
	#exit(0);
	
	my @list =();
	
	unless (defined $response->{'data'}) {
		die;
	}
	
	foreach my $node (@{$response->{'data'}}) {
		#print $node->{'id'}."\n";
		push(@list, $node->{'id'});
	}
	
	print "found ".@list. " nodes that can be deleted\n" if $debug;
	
	foreach my $node (@list) {
		my $ret = $shock->delete_node($node);
		defined($ret) or die;
		pprint_json($ret);
	}
	
	
	exit(0);
}

sub pprint_json {
    my ($data) = @_;
   # print STDOUT $shock->json->pretty->encode($data);
	print $shock->pretty($data);
}
