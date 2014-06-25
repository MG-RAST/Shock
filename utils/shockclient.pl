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
'synopsis' => 'shockclient.pl --show=<nodeid>',
'examples' => 'shockclient.pl --upload *.fasta',
'authors' => 'Wolfgang Gerlach',
'options' => [
	'',
	'Actions:',
	[ 'show=s'						, ""],
	[ 'upload'						, "upload files to Shock"],
	[ 'delete=s'					, "delete Shock node"],
	[ 'query=s'						, "querystring, e.g. key=value,key2=value2"],
	[ 'querynode=s'					, "querystring, e.g. key=value,key2=value2"],
	[ 'download=s'					, ""],
	[ 'makepublic=s'				, "make node public"],
	[ 'clean_tmp'					, ""],
	'',
	'Options:',
#	[ 'xx=s'						, "xx"],
	[ 'public'						, "uploaded files will be public (default private)"],
	[ 'attributes_string=s'         , "string containing attributes to be uploaded with file, must be valid JSON" ],
	[ 'attributes_file=s'           , "file containing attributes to be uploaded with file, must be valid JSON" ],
	[ 'url=s' 						, "url to Shock server (default $shockurl)" ],
	[ 'token=s' 					, "default from \$KB_AUTH_TOKEN" ],
	[ 'id_list'                     , "return nodes ids and not content for query (defualt off)" ],
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


my $value = undef;

if (defined($value = $h->{"query"})) {
	
	my @queries = split(/,|\=/, $value);
	my $response =  $shock->query(@queries);
	
	if (defined($h->{"id_list"})) {
	    my @nodes = ();
    	foreach my $node_obj (@{$response->{'data'}}) {
    	    print $node_obj->{'id'}."\n";
    	}
	} else {
	    pprint_json($response);
	}
	
	exit(0);
} elsif (defined($value = $h->{"querynode"})) {
	
	my @queries = split(/,|\=/, $value);
	my $response =  $shock->querynode(@queries);
	
	if (defined($h->{"id_list"})) {
	    my @nodes = ();
    	foreach my $node_obj (@{$response->{'data'}}) {
    	    print $node_obj->{'id'}."\n";
    	}
	} else {
	    pprint_json($response);
	}
	
	exit(0);
} elsif (defined($value = $h->{"delete"})) {
	
	
	my @nodes = split(',', $value);
	
	
	foreach my $node (@nodes) {
		my $response =  $shock->delete_node($node);
		pprint_json($response);
	}
	
	
	
	exit(0);
} elsif (defined($h->{"upload"})) {
	
	
	my @files = @ARGV;
	my $attr = {};
	if (defined($h->{"attributes_string"})) {
	    $attr = $shock->json->decode($h->{"attributes_string"});
	} elsif (defined($h->{"attributes_file"}) && (-s $h->{"attributes_file"})) {
	    open(JSON, $h->{"attributes_file"}) or die $!;
        my $json_str = do { local $/; <JSON> };
        close(JSON);
        $attr = $shock->json->decode($json_str);
	}
	
	foreach my $file (@files) {
		
		print "uploading ".$file."...\n" if $debug;
		
		my $shock_node = $attr ? $shock->upload('file' => $file, 'attr' => $attr) : $shock->upload('file' => $file);
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
		print $file." saved with id $id\n" if $debug;
		
		if (defined $h->{"public"}) {
			print "make id $id public...\n" if $debug;
			$shock->permisson_readable($id);
		}
	}
	
	exit(0);
} elsif (defined($value = $h->{"makepublic"})) {
	print "make id $value public...\n" if $debug;
	$shock->permisson_readable($value);
	exit(0);
} elsif (defined($value = $h->{"show"})) {
	
	
	my @nodes = split(',', $value);
	
		
	foreach my $node (@nodes) {
		my $response =  $shock->get('node/'.$node);
		pprint_json($response);
	}
	
	
	
	exit(0);
} elsif (defined($value = $h->{"download"})) {
	
	
	my @nodes = split(',', $value);
	
	
	foreach my $node (@nodes) {
		
		my $view_response =  $shock->get('node/'.$node);
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
		    print "$node downloaded to $filename\n" if $debug;
		}
	}
	
	
	
	exit(0);
	
} elsif (defined($h->{"clean_tmp"})) {
	
	my $shock = new SHOCK::Client($shockurl, $shocktoken);
	unless (defined $shock) {
		die;
	}
	
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
    print STDOUT $shock->json->pretty->encode($data);
}
