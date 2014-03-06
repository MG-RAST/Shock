#!/usr/bin/env perl

use strict;
use warnings;

use lib ".";

use Data::Dumper;
use SHOCK::Client;

eval "use USAGEPOD qw(parse_options); 1"
or die "module USAGEPOD.pm required: wget https://raw.github.com/MG-RAST/MG-RAST-Tools/master/tools/lib/USAGEPOD.pm";


my $shockurl = $ENV{'SHOCK_SERVER_URL'} || die "SHOCK_SERVER_URL not defined";

my $shocktoken=$ENV{'GLOBUSONLINE'} || $ENV{'KB_AUTH_TOKEN'};


sub shock_upload {
	my ($shock) = shift(@_);
	my @other_args = @_;
	
	my $shock_data = $shock->upload(@other_args); # "test.txt"
	unless (defined $shock_data) {
		die;
	}
	#print Dumper($shock_data);
	unless (defined $shock_data->{'id'}) {
		die;
	}
	
	return $shock_data->{id};
}

#######################################


my ($h, $help_text) = &parse_options (
'name' => 'shockclient.pl',
'version' => '1',
'synopsis' => 'shockclient.pl --show=<nodeid>',
'examples' => 'ls',
'authors' => 'Wolfgang Gerlach',
'options' => [
	'',
	'Actions:',
	[ 'show=s'						, ""],
	[ 'delete=s'					, ""],
	[ 'query=s'						, ""],
	[ 'download=s'					, ""],
	[ 'clean_tmp'					, ""],
#	'',
#	'Options:',
#	[ 'xx=s'						, "xx"],
	[ 'help|h'						, "", { hidden => 1  }]
	]
);



if ($h->{'help'} || keys(%$h)==0) {
	print $help_text;
	exit(0);
}


print "connect to SHOCK\n";
my $shock = new SHOCK::Client($shockurl, $shocktoken); # shock production
unless (defined $shock) {
	die;
}


my $value = undef;

if (defined($value = $h->{"query"})) {
	
	
	my @queries = split(',', $value);
	
	
	my $response =  $shock->query(@queries);
	print Dumper($response);
	
	my @nodes = ();
	foreach my $node_obj (@{$response->{'data'}}) {
		push(@nodes, $node_obj->{'id'});
	}
	
	print "nodes: ".join(',',@nodes)."\n";
	
	exit(0);
} elsif (defined($value = $h->{"delete"})) {
	
	
	my @nodes = split(',', $value);
	
	
	foreach my $node (@nodes) {
		my $response =  $shock->delete_node($node);
		print Dumper($response);
	}
	
	
	
	exit(0);
} elsif (defined($value = $h->{"show"})) {
	
	
	my @nodes = split(',', $value);
	
		
	foreach my $node (@nodes) {
		my $response =  $shock->get('node/'.$node);
		print Dumper($response);
	}
	
	
	
	exit(0);
} elsif (defined($value = $h->{"download"})) {
	
	
	my @nodes = split(',', $value);
	
	
	foreach my $node (@nodes) {
		
		my $view_response =  $shock->get('node/'.$node);
		print Dumper($view_response);
		#exit(0);
		
		my $filename  = $view_response->{'data'}->{'file'}->{'name'};
		unless (defined $filename) {
			die "filename not defined, cannot save.";
		}
		
		if (-e $filename) {
			die "file \"$filename\" already exists";
		}
		
		
		my $response =  $shock->download_to_path($node, $filename);
		print Dumper($response);
	}
	
	
	
	exit(0);
	
} elsif (defined($h->{"clean_tmp"})) {
	
	my $shock = new SHOCK::Client($shockurl, $shocktoken);
	unless (defined $shock) {
		die;
	}
	
	my $response =  $shock->query('temporary' => 1);
	
	#my $response =  $shock->query('statistics.length_max' => 1175);
	print Dumper($response);
	#exit(0);
	
	my @list =();
	
	unless (defined $response->{'data'}) {
		die;
	}
	
	foreach my $node (@{$response->{'data'}}) {
		#print $node->{'id'}."\n";
		push(@list, $node->{'id'});
	}
	
	print "found ".@list. " nodes that can be deleted\n";
	
	foreach my $node (@list) {
		my $ret = $shock->delete_node($node);
		defined($ret) or die;
		print Dumper($ret);
	}
	
	
	exit(0);
}
