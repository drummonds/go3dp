float boxframe3p9062513p7062501914p1062498090p026562501(vec3 p) {
float e=0.026562501;
vec3 b=vec3(1.899999976,6.800000191,1.999999881);
p = abs(p)-b;
vec3 q = abs(p+e)-e;
return min(min(
      length(max(vec3(p.x,q.y,q.z),0.0))+min(max(p.x,max(q.y,q.z)),0.0),
      length(max(vec3(q.x,p.y,q.z),0.0))+min(max(q.x,max(p.y,q.z)),0.0)),
      length(max(vec3(q.x,q.y,p.z),0.0))+min(max(q.x,max(q.y,p.z)),0.0));
}

float translate0p0p0p_boxframe3p9062513p7062501914p1062498090p026562501(vec3 p) {
vec3 t=vec3(0.,0.,0.);
return boxframe3p9062513p7062501914p1062498090p026562501(p-t);
}

float box3p79999995213p6000003814p0p101gqj0b17a9pm1(vec3 p) {
float r=0.100000001;
vec3 d=vec3(1.899999976,6.800000191,2.);
vec3 q = abs(p)-d+r;
return length(max(q,0.0)) + min(max(q.x,max(q.y,q.z)),0.0)-r;
}

float union_box3p79999995213p6000003814p0p101gqj0b17a9pm1_translate0p0p0p_boxframe3p9062513p7062501914p1062498090p026562501(vec3 p) {
return min(box3p79999995213p6000003814p0p101gqj0b17a9pm1(p),translate0p0p0p_boxframe3p9062513p7062501914p1062498090p026562501(p));
}

float translate0p0p0p_union_box3p79999995213p6000003814p0p101gqj0b17a9pm1_translate0p0p0p_boxframe3p9062513p7062501914p1062498090p026562501(vec3 p) {
vec3 t=vec3(0.,0.,0.);
return union_box3p79999995213p6000003814p0p101gqj0b17a9pm1_translate0p0p0p_boxframe3p9062513p7062501914p1062498090p026562501(p-t);
}

float scale_translate0p0p0p_union_box3p79999995213p6000003814p0p101gqj0b17a9pm1_translate0p0p0p_boxframe3p9062513p7062501914p1062498090p026562501(vec3 p) {
float s=0.036764704;
return translate0p0p0p_union_box3p79999995213p6000003814p0p101gqj0b17a9pm1_translate0p0p0p_boxframe3p9062513p7062501914p1062498090p026562501(p/s)*s;
}


float sdf(vec3 p) { return scale_translate0p0p0p_union_box3p79999995213p6000003814p0p101gqj0b17a9pm1_translate0p0p0p_boxframe3p9062513p7062501914p1062498090p026562501(p); }

// The MIT License
// Copyright © 2023 Inigo Quilez
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions: The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software. THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Distance to a vesica segment with 3 square roots, and to
// a vertival vesica segment with 2 square roots.
//
// 2D version here: https://www.shadertoy.com/view/cs2yzG
//
// List of other 3D SDFs:
//    https://www.shadertoy.com/playlist/43cXRl
// and 
//     https://iquilezles.org/articles/distfunctions


// https://iquilezles.org/articles/normalsSDF
vec3 calcNormal( in vec3 pos )
{
    vec2 e = vec2(1.0,-1.0)*0.5773;
    const float eps = 0.0005;
    return normalize( e.xyy*sdf( pos + e.xyy*eps ) + 
					  e.yyx*sdf( pos + e.yyx*eps ) + 
					  e.yxy*sdf( pos + e.yxy*eps ) + 
					  e.xxx*sdf( pos + e.xxx*eps ) );
}

// Antialiasing.
#define AA 3

void mainImage( out vec4 fragColor, in vec2 fragCoord )
{
     // camera movement	
	float an = 0.2*sin(iTime);
	vec3 ro = vec3( 1.0*sin(an), 0.4, 1.0*cos(an) );
    vec3 ta = vec3( 0.0, 0.0, 0.0 );
    // camera matrix
    vec3 ww = normalize( ta - ro );
    vec3 uu = normalize( cross(ww,vec3(0.0,1.0,0.0) ) );
    vec3 vv = normalize( cross(uu,ww));

        
    vec3 tot = vec3(0.0);
    
    #if AA>1
    for( int m=0; m<AA; m++ )
    for( int n=0; n<AA; n++ )
    {
        // pixel coordinates
        vec2 o = vec2(float(m),float(n)) / float(AA) - 0.5;
        vec2 p = (2.0*(fragCoord+o)-iResolution.xy)/iResolution.y;
        #else    
        vec2 p = (2.0*fragCoord-iResolution.xy)/iResolution.y;
        #endif

	    // create view ray
        vec3 rd = normalize( p.x*uu + p.y*vv + 1.5*ww );

        // raymarch
        const float tmax = 3.0;
        float t = 0.0;
        for( int i=0; i<256; i++ )
        {
            vec3 pos = ro + t*rd;
            float h = sdf(pos);
            if( h<0.0001 || t>tmax ) break;
            t += h;
        }
        
    
        // shading/lighting	
        vec3 col = vec3(0.0);
        if( t<tmax )
        {
            vec3 pos = ro + t*rd;
            vec3 nor = calcNormal(pos);
            float dif = clamp( dot(nor,vec3(0.57703)), 0.0, 1.0 );
            float amb = 0.5 + 0.5*dot(nor,vec3(0.0,1.0,0.0));
            col = vec3(0.2,0.3,0.4)*amb + vec3(0.8,0.7,0.5)*dif;
        }

        // gamma        
        col = sqrt( col );
	    tot += col;
    #if AA>1
    }
    tot /= float(AA*AA);
    #endif

	fragColor = vec4( tot, 1.0 );
}