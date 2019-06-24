#!/usr/bin/env perl

use strict;
use warnings;

open my $fh, "<", "main.go" || die $!;

my $version;
while (<$fh>) {
  $version = $1 if /^const\s+currentVersion.*?"([\d\.]+)"/;
}
close $fh;

die "no version?" unless defined $version;

# so lazy
system "rm", "-rf", "release", "dist";
system "mkdir", "release";
system "mkdir", "dist";

my %build = (
  win   => { env => { GOOS => 'windows', GOARCH => '386' }, filename => 'slicerdicer.exe' },
  linux => { env => { GOOS => 'linux',   GOARCH => '386' }, filename => 'slicerdicer' },
  mac   => { env => { GOOS => 'darwin',  GOARCH => '386' }, filename => 'slicerdicer' },
); 

foreach my $type (keys %build) {
  mkdir "release/$type";
}

foreach my $type (keys %build) {
  local $ENV{GOOS}   = $build{$type}->{env}->{GOOS};
  local $ENV{GOARCH} = $build{$type}->{env}->{GOARCH};
  system "go", "build", "-o", "release/$type/" . $build{$type}->{filename};
  system "zip", "-j", "dist/slicerdicer-$type-$version.zip", ( glob "release/$type/*" );
}

