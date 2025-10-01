import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { ArrowRight, ArrowLeft, Check, Globe, Server, Eye, AlertTriangle } from 'lucide-react';
import { ThemeProvider } from '@/components/theme-provider';

const INTRO_COOKIE_NAME = 'has_seen_intro';

interface Statistics {
  results: number;
  domain_stats: {
    unique_apex_domains: number;
    total_subdomains: number;
    total_domains: number;
  };
  ip_stats: {
    unique_ips: number;
  };
  target_info: {
    company_name: string;
    main_domain: string;
    scan_start_time: string;
  };
}

const Intro = () => {
  const [currentSlide, setCurrentSlide] = useState(0);
  const [statistics, setStatistics] = useState<Statistics | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    // Fetch statistics for intro content
    fetch('/api/statistics')
      .then(response => response.json())
      .then(data => setStatistics(data))
      .catch(error => console.error('Failed to fetch statistics:', error));
  }, []);

  const getIntroSlides = () => [
    {
      title: 'Welcome to Defend Denmark ASM',
      description: 'Attack Surface Mapping Tool',
      content: (
        <div className="space-y-4">
          <div className="flex items-center justify-center mb-6">
            <div className="rounded-lg bg-white p-6 shadow-sm border">
              <img src="/logo_red.png" alt="Defend Denmark Logo" className="h-16 w-16" />
            </div>
          </div>
          <p className="text-lg text-muted-foreground text-center">
            This is the Defend Denmark Attack Surface Mapping tool used to map out the 
            attack surface for <strong>{statistics?.target_info?.company_name || 'the target company'} </strong> 
            from the main domain <strong>{statistics?.target_info?.main_domain || 'N/A'}</strong>.
          </p>
          <p className="text-muted-foreground text-center">
            The scan was completed using passive enumeration techniques to provide insights 
            into what the company is exposing on the internet.
          </p>
        </div>
      ),
    },
    {
      title: 'Scan Results Overview',
      description: 'What we discovered',
      content: (
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center p-4 bg-muted/50 rounded-lg">
              <div className="flex items-center justify-center mb-2">
                <Eye className="h-8 w-8 text-purple-600" />
              </div>
              <div className="text-2xl font-bold">{statistics?.domain_stats?.unique_apex_domains || 0}</div>
              <div className="text-sm text-muted-foreground">Apex Domains</div>
            </div>
            <div className="text-center p-4 bg-muted/50 rounded-lg">
              <div className="flex items-center justify-center mb-2">
                <Globe className="h-8 w-8 text-orange-600" />
              </div>
              <div className="text-2xl font-bold">{statistics?.domain_stats?.total_subdomains || 0}</div>
              <div className="text-sm text-muted-foreground">Subdomains</div>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="text-center p-4 bg-muted/50 rounded-lg">
              <div className="flex items-center justify-center mb-2">
                <Globe className="h-8 w-8 text-blue-600" />
              </div>
              <div className="text-2xl font-bold">{statistics?.domain_stats?.total_domains || 0}</div>
              <div className="text-sm text-muted-foreground">Total Domains</div>
            </div>
            <div className="text-center p-4 bg-muted/50 rounded-lg">
              <div className="flex items-center justify-center mb-2">
                <Server className="h-8 w-8 text-green-600" />
              </div>
              <div className="text-2xl font-bold">{statistics?.ip_stats?.unique_ips || 0}</div>
              <div className="text-sm text-muted-foreground">Unique IPs</div>
            </div>
          </div>
          <p className="text-sm text-muted-foreground text-center mt-4">
            These numbers represent the digital footprint discovered through passive reconnaissance.
          </p>
        </div>
      ),
    },
    {
      title: 'Important Disclaimer',
      description: 'Passive reconnaissance only',
      content: (
        <div className="space-y-4">
          <div className="flex items-center justify-center mb-4">
            <div className="rounded-full bg-yellow-50 p-4">
              <AlertTriangle className="h-12 w-12 text-yellow-600" />
            </div>
          </div>
          <div className="space-y-3">
            <div className="flex items-start gap-3">
              <div className="rounded-full bg-green-100 p-2 mt-1">
                <Check className="h-4 w-4 text-green-600" />
              </div>
              <div>
                <h4 className="font-medium">Passive Enumeration Only</h4>
                <p className="text-sm text-muted-foreground">No exploitation attempts were made during this scan</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="rounded-full bg-green-100 p-2 mt-1">
                <Check className="h-4 w-4 text-green-600" />
              </div>
              <div>
                <h4 className="font-medium">Public Information</h4>
                <p className="text-sm text-muted-foreground">All data gathered from publicly available sources</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="rounded-full bg-green-100 p-2 mt-1">
                <Check className="h-4 w-4 text-green-600" />
              </div>
              <div>
                <h4 className="font-medium">Insight Purpose</h4>
                <p className="text-sm text-muted-foreground">Provides visibility into exposed assets and attack surface</p>
              </div>
            </div>
          </div>
          <p className="text-sm text-muted-foreground text-center mt-4 p-3 bg-blue-50 rounded-lg">
            This tool helps organizations understand their external-facing assets for security assessment purposes.
          </p>
        </div>
      ),
    },
    {
      title: 'Get Started',
      description: 'Explore your attack surface',
      content: (
        <div className="space-y-4">
          <p className="text-lg text-muted-foreground text-center">
            Ready to explore the discovered attack surface? Use the navigation menu to access different views:
          </p>
          <ul className="space-y-2 text-muted-foreground">
            <li>• <strong>Dashboard</strong> - Overview and key statistics</li>
            <li>• <strong>Gallery</strong> - Visual screenshots of discovered services</li>
            <li>• <strong>Overview</strong> - Detailed table view of all results</li>
            <li>• <strong>IPs</strong> - Browse discovered services by IP address</li>
            <li>• <strong>Domains</strong> - Browse discovered services by domain</li>
          </ul>
          <p className="text-sm text-muted-foreground text-center mt-4">
            Click "Get Started" to begin exploring the attack surface mapping results.
          </p>
        </div>
      ),
    },
  ];

  const introSlides = getIntroSlides();
  const isLastSlide = currentSlide === introSlides.length - 1;
  const isFirstSlide = currentSlide === 0;

  const handleNext = () => {
    if (isLastSlide) {
      // Set the cookie and navigate to dashboard
      document.cookie = `${INTRO_COOKIE_NAME}=true; path=/; max-age=31536000`; // 1 year
      navigate('/');
    } else {
      setCurrentSlide(currentSlide + 1);
    }
  };

  const handlePrevious = () => {
    if (!isFirstSlide) {
      setCurrentSlide(currentSlide - 1);
    }
  };

  const handleSkip = () => {
    document.cookie = `${INTRO_COOKIE_NAME}=true; path=/; max-age=31536000`; // 1 year
    navigate('/');
  };

  const slide = introSlides[currentSlide];

  return (
    <ThemeProvider defaultTheme="light" storageKey="ui-theme">
      <div className="flex min-h-screen w-full items-center justify-center bg-background p-4">
        <Card className="w-full max-w-2xl">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <CardTitle className="text-3xl">{slide.title}</CardTitle>
                <CardDescription className="text-base mt-2">{slide.description}</CardDescription>
              </div>
              <Button variant="ghost" size="sm" onClick={handleSkip}>
                Skip
              </Button>
            </div>
          </CardHeader>
          <CardContent className="min-h-[300px]">
            {slide.content}
          </CardContent>
          <CardFooter className="flex items-center justify-between">
            <div className="flex gap-2">
              {introSlides.map((_, index) => (
                <div
                  key={index}
                  className={`h-2 w-2 rounded-full transition-all ${
                    index === currentSlide
                      ? 'bg-primary w-8'
                      : 'bg-muted'
                  }`}
                />
              ))}
            </div>
            <div className="flex gap-2">
              {!isFirstSlide && (
                <Button
                  variant="outline"
                  onClick={handlePrevious}
                >
                  <ArrowLeft className="h-4 w-4 mr-2" />
                  Previous
                </Button>
              )}
              <Button onClick={handleNext}>
                {isLastSlide ? (
                  <>
                    Get Started
                    <Check className="h-4 w-4 ml-2" />
                  </>
                ) : (
                  <>
                    Next
                    <ArrowRight className="h-4 w-4 ml-2" />
                  </>
                )}
              </Button>
            </div>
          </CardFooter>
        </Card>
      </div>
    </ThemeProvider>
  );
};

export default Intro;